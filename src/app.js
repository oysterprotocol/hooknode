import bodyParser from "body-parser";
import compression from "compression";
import dotenv from "dotenv";
import errorHandler from "errorhandler";
import express from "express";
import morgan from "morgan";

import routes from "./routes";

// Configure Environment
dotenv.config();

// Create Express server
const app = express();

// Express configuration
app.set("port", process.env.PORT || 3000);
app.use(compression());
app.use(morgan("dev"));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));
if (app.get("env") !== "production") {
  app.use(errorHandler());
}

routes(app);

export default app;
