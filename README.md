# YanPlatform

YanPlatform is a continuous intelligence engine designed to map chokepoints, assess risks, and simulate alternative reroutes for critical mineral supply chains. The platform specializes in tracking highly concentrated, strategically important resources (starting with Gallium and Germanium) to provide proactive geopolitical risk profiling.

## Architecture Architecture

YanPlatform is composed of two main components:
*   **Backend (`/backend`)**: A high-performance Go-based API server that handles data ingestion, risk scoring, and shadow reroute computations.
*   **Frontend (`/frontend`)**: A Flutter-based dashboard application providing real-time visualizations of supply chains, chokepoints, and risk trends.

## Core Features

1.  **Risk Assessment Engine**: Computes dynamic risk scores for specific regions based on:
    *   Supply Concentration (40%)
    *   Geopolitical Tension (30%)
    *   Trade Policy Signals (20%)
    *   Logistics Risk (10%)
2.  **Shadow Rerouting Simulation**: Automatically identifies and ranks alternative suppliers whenever a designated "high-risk" region exceeds its risk threshold.
3.  **Real-Time Data Pipelines**:
    *   **GDELT Integration**: Ingests geopolitical events and policy changes via BigQuery.
    *   **UN Comtrade API**: Pulls real-time international trade flows and metrics.
    *   **NVIDIA NIM**: Provides advanced LLM-based sentiment classification on geopolitical events.
4.  **Flexible Data Layer**: Operates entirely in-memory for testing and development, with seamless toggling to a live cloud database (Google Cloud Firestore) using `USE_FIRESTORE=true`.

## Backend Setup (Go)

### Prerequisites
*   Go 1.22+
*   (Optional) Firebase/Firestore credentials for cloud storage
*   (Optional) API Keys for UN Comtrade and NVIDIA NIM for live data ingestion

### Running the Backend

Navigate to the backend directory:
```bash
cd backend
```

Run the server locally:
```bash
go run cmd/server/main.go
```
The API server will run by default on `http://localhost:8080`.

### Configuration
Configuration is managed via `config.json` and overrides via Environment Variables:
*   `PORT`: Override the server port.
*   `USE_FIRESTORE`: Set to `true` to switch the datastore from in-memory (`MemoryStore`) to Google Cloud Firestore (`FirestoreStore`).
*   `FIREBASE_PROJECT_ID`: Your GCP project ID.
*   `GOOGLE_APPLICATION_CREDENTIALS`: Path to your service account key.
*   `NIM_API_KEY`: API key for NVIDIA NIM sentiment analysis.
*   `COMTRADE_API_KEY`: API key for the UN Comtrade trade flow pipeline.

## Frontend Setup (Flutter)

Navigate to the frontend directory:
```bash
cd frontend
```

Fetch dependencies and start the application:
```bash
flutter pub get
flutter run -d chrome
```

## Current Status

YanPlatform has successfully completed its localized, in-memory Minimum Viable Product (Phase 1). It is currently evolving into **Phase 2**, which integrates live, cloud-based infrastructure (Firestore) and is expanding to map multi-resource clusters including the "EV Battery Belt" (Lithium, Cobalt, and Graphite).

## License

Copyright 2026. All rights reserved.
