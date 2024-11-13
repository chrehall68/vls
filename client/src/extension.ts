import { ExtensionContext, ExtensionMode, workspace } from "vscode";

import { createWriteStream, existsSync, mkdirSync } from "fs";
import { chmod } from "fs/promises";
import * as net from "net";
import { sep } from "path";
import { Readable } from "stream";
import { finished } from "stream/promises";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  StreamInfo,
} from "vscode-languageclient/node";

let client: LanguageClient;

async function downloadToBin(
  ctx: ExtensionContext,
  url: string,
  filename: string
) {
  const res = await fetch(url);
  if (!existsSync(ctx.asAbsolutePath("bin"))) {
    mkdirSync(ctx.asAbsolutePath("bin"), { recursive: true });
  }
  const actualFileName = ctx.asAbsolutePath(`bin${sep}${filename}`);
  const fileStream = createWriteStream(actualFileName, { flags: "wx" });
  await finished(Readable.fromWeb(res.body).pipe(fileStream));
}

async function resolveServerExecutable(ctx: ExtensionContext): Promise<string> {
  const version = "1.0.3";
  const platformDetails = {
    win32: {
      url: `https://github.com/chrehall68/vls/releases/download/${version}/vls-windows-amd64.exe`,
      filename: "vls.exe",
      doChmod: false,
    },
    darwin: {
      url: `https://github.com/chrehall68/vls/releases/download/${version}/vls-macos-amd64`,
      filename: "vls",
      doChmod: true,
    },
    linux: {
      url: `https://github.com/chrehall68/vls/releases/download/${version}/vls-linux-amd64`,
      filename: "vls",
      doChmod: true,
    },
  };

  const platform = process.platform;
  if (!platformDetails[platform]) {
    throw new Error(`Unsupported platform: ${platform}`);
  }
  const { url, filename, doChmod } = platformDetails[platform];
  if (!existsSync(ctx.asAbsolutePath(`bin${sep}${filename}`))) {
    await downloadToBin(ctx, url, filename);
    if (doChmod) {
      // make it executable; rx-rx-rx
      await chmod(ctx.asAbsolutePath(`bin${sep}${filename}`), 0o555);
    }
  }
  return ctx.asAbsolutePath(`bin${sep}${filename}`);
}

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

  const executable = await resolveServerExecutable(ctx);
  return {
    command: executable,
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
