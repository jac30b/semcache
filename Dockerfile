FROM chromadb/chroma:latest

# Set environment variables for persistence
ENV IS_PERSISTENT=TRUE
ENV PERSIST_DIRECTORY=/data
