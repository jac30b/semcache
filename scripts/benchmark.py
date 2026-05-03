import json
import requests
import time
import sys

# Configuration
URL = "http://localhost:8000/query"

def run_benchmark():
    try:
        with open("test_metadata.json", "r") as f:
            metadata = json.load(f)
    except Exception as e:
        print(f"Error loading metadata: {e}")
        return

    queries = metadata["queries"]
    total = len(queries)
    
    actual_hits = 0
    actual_misses = 0
    correct_predictions = 0
    errors = 0
    
    print(f"Starting benchmark for {total} queries...")
    print("-" * 50)
    
    # Track intents that have been seen during THIS run
    # The application storage might already be populated, 
    # so we should ideally start with a clean storage for a perfect comparison.
    # However, we will judge "HIT" based on response time or if we can find a way to detect it.
    # Looking at http.go, the server doesn't send a cache header.
    # We will use response time as a heuristic if needed, or just track results.
    
    # Actually, the server returns c.String for hits and LLM responses.
    # But for LLM responses, it uses c.JSON at the very end (wait, no, it's mixed).
    
    # Let's check http.go again.
    # exact match: return c.String(http.StatusOK, e.Answer)
    # nearest match: return c.String(http.StatusOK, entries[0].Answer)
    # ask llm: return c.JSON(http.StatusOK, e.Answer)
    
    # This is perfect! c.JSON will have quotes and potentially different content type, 
    # but c.String will be raw text.
    
    results = []
    
    for i, item in enumerate(queries):
        query_text = item["query"]
        expected = item["expected"]
        
        start_time = time.time()
        try:
            response = requests.get(URL, params={"query": query_text})
            duration = (time.time() - start_time) * 1000 # ms
            
            # Detection using the new X-Cache header
            actual = response.headers.get("X-Cache", "MISS")
            
            # Alternative: if it's the first time we see this intent in metadata it's a MISS, 
            # but if the cache was ALREADY populated from a previous run, it might be a HIT.
            
            if actual == "HIT":
                actual_hits += 1
            else:
                actual_misses += 1
                
            if actual == expected:
                correct_predictions += 1
            
            results.append({
                "query": query_text,
                "expected": expected,
                "actual": actual,
                "duration_ms": duration,
                "status": response.status_code
            })
            
        except Exception as e:
            print(f"Error on query {i}: {e}")
            errors += 1
            continue
            
        if (i + 1) % 50 == 0:
            print(f"Processed {i+1}/{total}...")

    # Summary
    print("-" * 50)
    print("Benchmark Results:")
    print(f"Total Queries: {total}")
    print(f"Actual Hits: {actual_hits}")
    print(f"Actual Misses: {actual_misses}")
    print(f"Hit Rate: {actual_hits/total:.2%}")
    print(f"Errors: {errors}")
    
    print("-" * 50)
    print("Comparison with Expectations:")
    print(f"Expected Hit Rate: {metadata['summary']['theoretical_hit_rate']:.2%}")
    print(f"Matches Expectation: {correct_predictions/total:.2%}")
    
    # Save results
    with open("benchmark_results.json", "w") as f:
        json.dump({
            "summary": {
                "total": total,
                "actual_hits": actual_hits,
                "actual_misses": actual_misses,
                "hit_rate": actual_hits/total,
                "expected_hit_rate": metadata['summary']['theoretical_hit_rate'],
                "accuracy_vs_metadata": correct_predictions/total
            },
            "details": results
        }, f, indent=2)
    print("Full results saved to benchmark_results.json")

if __name__ == "__main__":
    run_benchmark()
