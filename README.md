# Synaptica

Behavioral Analytics Platform for Public Tech Discourse.

Synaptica ingests public discussions from Hacker News, analyzes behavioral signals using rule-based NLP heuristics, and exposes insights through a modern analytics dashboard.

The project demonstrates end-to-end product engineering:

* Go backend
* PostgreSQL (Neon)
* Next.js frontend
* Data ingestion pipelines
* Behavioral analytics
* Emotion detection
* Trend discovery
* Dashboard visualization

---

## Architecture

```text
Hacker News
     ↓
 Ingestion
     ↓
 PostgreSQL
     ↓
 Analysis Engine
     ↓
 Analytics APIs
     ↓
 Next.js Dashboard
```

---

## Core Features

### Discussion Ingestion

* Hacker News story ingestion
* Recursive comment ingestion
* Keyword matching

### Behavioral Analysis

* Sentiment analysis
* Emotion detection
* Toxicity scoring
* Cognitive load scoring

### Analytics

* Emotion timeline
* Trending keywords
* Hot topics detection
* Post insights
* Dashboard overview

---

## Tech Stack

### Backend

* Go
* Chi Router
* sqlx
* PostgreSQL
* Neon

### Frontend

* Next.js
* React
* TypeScript
* Tailwind CSS
* Recharts

---

## Run Backend

```bash
go run ./cmd/api
```

Health Check:

```bash
curl http://localhost:8080/health
```

---

## Run Frontend

```bash
npm install
npm run dev
```

---

## Example APIs

Dashboard:

```bash
curl "http://localhost:8080/api/v1/dashboard/overview"
```

Trending Keywords:

```bash
curl "http://localhost:8080/api/v1/analytics/trending-keywords"
```

Hot Topics:

```bash
curl "http://localhost:8080/api/v1/analytics/hot-topics"
```

Post Insight:

```bash
curl "http://localhost:8080/api/v1/insights/posts/<post_id>"
```

---

## Why Synaptica?

Most analytics systems focus on engagement metrics.

Synaptica focuses on behavioral signals:

* What topics people discuss
* How emotions shift over time
* Which discussions become emotionally intense
* How cognitive complexity evolves within public discourse

The project was built as a portfolio-grade demonstration of Product Engineering, Data Engineering, and Behavioral Analytics.
