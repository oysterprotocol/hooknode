import * as iotaController from "./controllers/iota-controller";

export default app => {
  app.post("/HookListener.php", iotaController.legacy);
  app.get("/debug", iotaController.debug);

  // TODO: Mirror IRI API.

  return app;
};
