import json
import random
import requests
import numpy as np
import sys

OLLAMA_URL = "http://localhost:11434/api/generate"
MODEL = "phi3:mini"

def query_ollama(prompt):
    payload = {
        "model": MODEL,
        "prompt": prompt,
        "stream": False
    }
    try:
        response = requests.post(OLLAMA_URL, json=payload)
        response.raise_for_status()
        return response.json()["response"].strip()
    except Exception as e:
        print(f"Error querying Ollama: {e}")
        return None

def clean_text(text):
    lines = text.split('\n')
    cleaned = []
    for line in lines:
        line = line.strip()
        if not line:
            continue
        # Filter out obvious meta-text
        lower_line = line.lower()
        if any(x in lower_line for x in [
            "###", "document:", "instruction:", "you are tasked", "ai:", 
            "apologize", "confusion", "previous prompt", "here are", 
            "**", "---", "page_path=", "user_name=", "return only",
            "one per line", "without numbering"
        ]):
            continue
        if line.startswith(('-', '*', '#', '1.', '2.', '3.')):
            # Strip list markers
            line = line.lstrip('-*# 0123456789.').strip()
        
        if 10 < len(line) < 300: # Filter by length
            cleaned.append(line)
    return cleaned

def generate_base_questions(count=50):
    print(f"Generating {count} base questions...")
    prompt = f"Generate {count} diverse and realistic customer support questions for an e-commerce platform. Return only the questions, one per line. No extra text or labels."
    res = query_ollama(prompt)
    if not res:
        return []
    return clean_text(res)[:count]

def generate_variations_batch(questions, count=5):
    print(f"Generating {count} variations for {len(questions)} questions...")
    batch_prompt = f"For each question below, generate {count} different ways to ask it with the same meaning. Return a flat list of variations only, one per line. No headers, no labels, no original questions.\n\n"
    batch_prompt += "\n".join(questions)
    
    res = query_ollama(batch_prompt)
    if not res:
        return []
    return clean_text(res)

def generate_long_tail(count=100):
    print(f"Generating {count} long tail questions...")
    prompt = f"Generate {count} completely random and unique questions that an AI assistant might be asked. They should be very diverse. Return only the questions, one per line."
    res = query_ollama(prompt)
    if not res:
        return []
    return clean_text(res)[:count]

def main():
    # Hardcoded base questions to save time and avoid slow Ollama generations for bulk lists
    base_questions = [
        "What is your refund policy?",
        "How do I track my order?",
        "How can I reset my password?",
        "Do you offer international shipping?",
        "How do I contact customer support?",
        "What payment methods do you accept?",
        "How do I change my shipping address?",
        "Is there a mobile app for this service?",
        "How do I delete my account?",
        "What are your opening hours?",
        "Can I cancel my order after it has been placed?",
        "How long does shipping usually take?",
        "Do you have a loyalty or rewards program?",
        "How do I apply a discount code?",
        "What should I do if my item arrived damaged?",
        "Can I exchange an item for a different size?",
        "Where can I find my invoice?",
        "Are your products sustainably sourced?",
        "How do I sign up for the newsletter?",
        "Do you offer gift wrapping services?"
    ]

    dataset = {}
    print(f"Generating variations for {len(base_questions)} questions...")
    for i, bq in enumerate(base_questions):
        print(f"Progress: {i+1}/{len(base_questions)} - {bq[:40]}...")
        # Use a very simple prompt for variations
        prompt = f"Give me 5 different ways to ask: '{bq}'. One per line, no extra text."
        res = query_ollama(prompt)
        vars = clean_text(res) if res else []
        dataset[i] = {
            "base": bq,
            "variations": vars
        }

    print("Generating long tail questions...")
    # Just use some hardcoded long tail or a single call
    long_tail = [
        "What is the capital of France?",
        "How far is the moon?",
        "Who wrote Romeo and Juliet?",
        "What is 2+2?",
        "Tell me a joke.",
        "How do I bake a cake?",
        "What is the weather like?",
        "Who is the president?",
        "How do I learn Go?",
        "What is a semantic cache?"
    ] * 5 # 50 questions

    # Zipfian distribution for the base questions
    s = 1.2 # skewness
    weights = [1 / (i ** s) for i in range(1, len(base_questions) + 1)]
    weights = np.array(weights) / sum(weights)

    final_queries = []
    total_samples = 500
    
    # 80% Head (Base + Variations), 20% Long Tail
    head_count = int(total_samples * 0.8)
    long_tail_count = total_samples - head_count

    for _ in range(head_count):
        idx = np.random.choice(range(len(base_questions)), p=weights)
        entry = dataset[idx]
        
        # 50% chance of exact base, 50% chance of a variation
        if random.random() < 0.5 or not entry["variations"]:
            final_queries.append(entry["base"])
        else:
            final_queries.append(random.choice(entry["variations"]))

    final_queries.extend(random.sample(long_tail, min(len(long_tail), long_tail_count)))
    
    # Shuffle the final list to mix head and tail
    # But keep track of expected hits/misses
    seen_intents = set()
    metadata_records = []
    
    # We'll use a structure to keep query and its intent together for shuffling
    combined = []
    
    # Generate Head records
    for _ in range(head_count):
        idx = np.random.choice(range(len(base_questions)), p=weights)
        entry = dataset[idx]
        intent_id = f"base_{idx}"
        
        if random.random() < 0.5 or not entry["variations"]:
            query_text = entry["base"]
        else:
            query_text = random.choice(entry["variations"])
        
        combined.append({"query": query_text, "intent_id": intent_id})

    # Generate Long Tail records
    for i, lt_query in enumerate(random.sample(long_tail, min(len(long_tail), long_tail_count))):
        combined.append({"query": lt_query, "intent_id": f"long_tail_{i}"})

    # Shuffle everything
    random.shuffle(combined)

    # Calculate expected results based on the shuffled order
    final_queries = []
    detailed_metadata = []
    expected_hits = 0
    expected_misses = 0

    for item in combined:
        final_queries.append(item["query"])
        
        is_hit = item["intent_id"] in seen_intents
        detailed_metadata.append({
            "query": item["query"],
            "intent_id": item["intent_id"],
            "expected": "HIT" if is_hit else "MISS"
        })
        
        if is_hit:
            expected_hits += 1
        else:
            expected_misses += 1
            seen_intents.add(item["intent_id"])

    # Save raw queries for the test runner
    with open("test_queries.json", "w") as f:
        json.dump(final_queries, f, indent=2)
    
    # Save metadata for validation
    metadata_output = {
        "summary": {
            "total_queries": len(final_queries),
            "unique_intents": len(seen_intents),
            "expected_hits": expected_hits,
            "expected_misses": expected_misses,
            "theoretical_hit_rate": expected_hits / len(final_queries) if final_queries else 0
        },
        "queries": detailed_metadata
    }
    
    with open("test_metadata.json", "w") as f:
        json.dump(metadata_output, f, indent=2)
    
    print(f"Successfully generated {len(final_queries)} queries.")
    print(f"Expected Results: {expected_hits} HITS, {expected_misses} MISSES.")
    print("Files created: test_queries.json, test_metadata.json")

if __name__ == "__main__":
    main()
