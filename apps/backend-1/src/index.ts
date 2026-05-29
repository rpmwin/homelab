import express from "express";
import { collectDefaultMetrics, Registry, Counter, Histogram } from "prom-client";

const app = express();
const register = new Registry();
collectDefaultMetrics({ register });

const httpRequests = new Counter({
  name: "http_requests_total",
  help: "Total HTTP requests",
  labelNames: ["method", "route", "status"],
  registers: [register],
});

const httpDuration = new Histogram({
  name: "http_request_duration_ms",
  help: "HTTP request duration in ms",
  labelNames: ["method", "route"],
  registers: [register],
});

// 1. Add your custom logger middleware here
app.use((req, res, next) => {
  const start = Date.now();

  res.on("finish", () => {
    const duration = Date.now() - start;
    console.log(`${req.method} ${req.originalUrl} ${res.statusCode} - ${duration}ms`);
    httpRequests.inc({ method: req.method, route: req.path, status: res.statusCode });
    httpDuration.observe({ method: req.method, route: req.path }, duration);
  });

  next();
});

// Your existing routes remain exactly the same
app.get("/", (req, res) => {
  res.send("hello world");
});

app.get("/health", (req, res) => {
  res.send("Health check passed!");
});

app.get("/api", (req, res) => {
  res.json({ message: "Hello from the API!" });
});

app.get("/api/data", (req, res) => {
  const data = {
    id: 1,
    name: "Sample Data",
    description: "This is some sample data from the API.",
  };
  res.json(data);
});

app.get("/metrics", async (req, res) => {
  res.set("Content-Type", register.contentType);
  res.end(await register.metrics());
});

app.listen(3000, () => {
  console.log("Server is running on port 3000");
});
