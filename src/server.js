import app from "./app";

/**
 * Start Express server.
 */
app.listen(app.get("port"), () => {
  process.stdout.write(
    `Running on http://localhost:${app.get("port")} in ${app.get("env")}`,
  );
  process.stdout.write("Press CTRL-C to stop\n");
});
