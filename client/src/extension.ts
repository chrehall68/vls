import { ExtensionContext, ExtensionMode, workspace } from "vscode";

import * as net from "net";
import * as path from "path";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  StreamInfo,
} from "vscode-languageclient/node";

let client: LanguageClient;

async function getServerOptions(ctx: ExtensionContext): Promise<ServerOptions> {
  if (ctx.extensionMode == ExtensionMode.Development) {
    // In debug mode, the server is launched by VSCode (on this project) with Go debugger
    // We need to connect to it with a socket because it's not a child process, no easy way to get its stdin/stdout (on linux, reading /proc/<pid>/0 and 1 is doable, but that's not cross platform)
    const connectionInfo = {
      host: "localhost",
      port: 60256,
    };
    return () => {
      const sock = net.connect(connectionInfo);
      return Promise.resolve({
        reader: sock,
        writer: sock,
      } as StreamInfo);
    };
  }

  let filename = "verilog_language_server";
  if (process.platform == "win32") {
    filename += ".exe";
  }

  return {
    command: ctx.asAbsolutePath(path.join("bin", filename)),
    args: [],
  };
}

export async function activate(ctx: ExtensionContext) {
  const serverOptions = await getServerOptions(ctx);

  // Options to control the language client
  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      {
        language: "verilog",
        scheme: "file",
      },
    ],
    synchronize: {
      // Notify the server about file changes to '.clientrc files contained in the workspace
      fileEvents: workspace.createFileSystemWatcher("**/.clientrc"),
    },
  };

  // Create the language client and start the client.
  client = new LanguageClient(
    "verilogLS",
    "Verilog Language Server",
    serverOptions,
    clientOptions
  );

  // If serverOptions specified a node module or an executable, it will be started automatically
  // If serverOptions is a callback, it will be run to get a connection
  console.log("Starting client");
  client.start();
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}
