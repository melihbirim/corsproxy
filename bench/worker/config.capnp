using Workerd = import "/workerd/workerd.capnp";

const config :Workerd.Config = (
  services = [
    (name = "main", worker = .mainWorker),
    (name = "internet", network = (allow = ["public", "private"])),
  ],
  sockets = [ (name = "http", address = "*:8787", http = (), service = "main") ],
);

const mainWorker :Workerd.Worker = (
  modules = [ (name = "index.js", esModule = embed "src/index.js") ],
  compatibilityDate = "2026-10-01",
  globalOutbound = "internet",
);
