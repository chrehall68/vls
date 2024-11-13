import { spawn } from "child_process";
import * as path from "path";
import * as fs from "fs";
const fsp = fs.promises;
import * as process from "process";
import { fileURLToPath } from 'url';

function getExecutableFilename(basename) {
  if (process.platform == "win32") {
    return basename + ".exe";
  } else {
    return basename;
  }
}

const filename = fileURLToPath(import.meta.url);
const dirname = path.dirname(filename);

// cd to repository root
process.chdir(path.join(dirname, ".."));

fsp.copyFile("README.md", path.join("client", "README.md"));
fsp.copyFile("LICENSE", path.join("client", "LICENSE"));

const goBuild = spawn(
  "go",
  ["build", "-o", path.resolve("client", "bin", getExecutableFilename("verilog_language_server"))],
  { cwd: "server" },
);
goBuild.on('exit', exitCode => {
  if (exitCode != 0) {
    throw `go build failed with exit code ${exitCode}`;
  }
});
