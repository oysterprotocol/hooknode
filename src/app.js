import bodyParser from "body-parser";
import compression from "compression";
import errorHandler from "errorhandler";
import express from "express";
import morgan from "morgan";

// Create Express server
const app = express();

// Express configuration
app.set("port", process.env.PORT || 8080);
app.use(compression());
app.use(morgan("dev"));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));
if (app.get("env") !== "production") {
  app.use(errorHandler());
}

export default app;
