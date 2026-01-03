# Distributed MapReduce Word Count

A distributed MapReduce implementation for word counting, built with Go, Docker, AWS ECS Fargate, and S3.

## 🚀 Overview

This project demonstrates a cloud-based MapReduce framework that processes large text files in parallel using distributed computing. The implementation achieves a **2.07x speedup** through parallelization.

## 📊 Performance Results

- **Average Execution Time**: 1.071 seconds
- **Speedup Factor**: 2.07x compared to sequential processing
- **Consistency**: Stable performance across multiple runs (< 600ms variance)

### Phase Breakdown
- **Split Phase**: 268ms (25% of total time)
- **Map Phase**: 572ms (53% of total time) - Benefits most from parallelization
- **Reduce Phase**: 231ms (22% of total time)

## 🏗️ Architecture

The system consists of four main components:

1. **Splitter Service**: Divides input file into 3 equal chunks
2. **Mapper Services** (3 instances): Process chunks in parallel, counting words
3. **Reducer Service**: Aggregates results from all mappers
4. **S3 Storage**: Stores input, intermediate, and final results

```
Input File (S3)
      ↓
   Splitter → 3 Chunks (S3)
      ↓
   Mappers (parallel) → Intermediate Results (S3)
      ↓
   Reducer → Final Word Count (S3)
```

## 🛠️ Technology Stack

- **Language**: Go
- **Containerization**: Docker
- **Orchestration**: AWS ECS Fargate
- **Storage**: AWS S3
- **Registry**: AWS ECR

## 📁 Project Structure

```
.
├── splitter/
│   ├── main.go          # Splitter service implementation
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── mapper/
│   ├── main.go          # Mapper service implementation
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── reducer/
│   ├── main.go          # Reducer service implementation
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── experiment.go        # Performance testing script
├── test_performance.go  # Basic performance test
├── performance_chart.html       # Performance visualization
├── results_visualization.html   # Detailed results dashboard
├── Report.txt          # Performance analysis report
└── input.txt           # Sample input file
```

## 🚦 Getting Started

### Prerequisites
- Go 1.x
- Docker
- AWS Account with ECS, S3, and ECR access
- AWS CLI configured

### Local Testing

1. **Clone the repository**
```bash
git clone <your-repo-url>
cd mapreduce-wordcount
```

2. **Build services**
```bash
# Build splitter
cd splitter && docker build -t splitter .

# Build mapper
cd ../mapper && docker build -t mapper .

# Build reducer
cd ../reducer && docker build -t reducer .
```

3. **Run performance tests**
```bash
go run experiment.go
```

## 📈 API Endpoints

### Splitter Service
- `GET /split?bucket=<bucket>&file=<filename>`
- Returns: JSON with chunk URLs

### Mapper Service
- `GET /map?chunk_url=<s3-url>`
- Returns: JSON with mapped result URL

### Reducer Service
- `GET /reduce`
- Returns: JSON with final result URL

## 🔍 Key Features

- **Parallel Processing**: Multiple mapper instances process data simultaneously
- **Cloud-Native**: Built for AWS with containerized microservices
- **Scalable**: Easy to add more mapper instances for larger datasets
- **Reliable**: Consistent performance with minimal variance
- **Storage-Efficient**: Uses S3 for intermediate and final results

## 📊 Performance Testing

Run the experiment to see consistency across 5 runs:

```bash
go run experiment.go
```

This will output:
- Individual run times for each phase
- Average execution times
- Speedup calculations
- Performance metrics

## 🎯 Results Visualization

Open `results_visualization.html` in a browser to see:
- Performance consistency graphs
- Phase distribution charts
- Sequential vs. Parallel comparison
- Speedup analysis

