import { spawn } from "child_process";
import * as path from "path";
import * as fs from "fs";
const fsp = fs.promises;
import * as process from "process";
import { fileURLToPath } from 'url';

function getExecutableFilename(basename) {
  if (process.env["GOOS"] == "windows") {
    return basename + ".exe";
  } else {
    return basename;
  }
}

const filename = fileURLToPath(import.meta.url);
const dirname = path.dirname(filename);

// cd to repository root
process.chdir(path.join(dirname, ".."));

// Misc resource files
fsp.copyFile("README.md", path.join("client", "README.md"));
fsp.copyFile("LICENSE", path.join("client", "LICENSE"));

const goBuild = spawn(
  "go",
  ["build", "-o", path.resolve("client", "bin", getExecutableFilename("verilog_language_server"))],
  // Target OS and architecture should be specified to the node process running this script, e.g.
  // bash -c 'GOOS=windows GOARCH=amd64 node build.mjs'
  { cwd: "server" },
);
goBuild.on('exit', exitCode => {
  if (exitCode != 0) {
    throw `go build failed with exit code ${exitCode}`;
  }
});
