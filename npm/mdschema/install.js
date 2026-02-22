"use strict";

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const { getBinaryPath, getGoReleaserTarget } = require("./lib/platform");

// Check if binary already exists from optionalDependencies
try {
  const existing = getBinaryPath();
  if (fs.existsSync(existing)) {
    process.exit(0);
  }
} catch {
  // Expected when platform package not installed — continue to download
}

const pkg = require("./package.json");
const version = pkg.version;

const target = getGoReleaserTarget();
const ext = process.platform === "win32" ? ".exe" : "";
const archiveName = `mdschema_${version}_${target.os}_${target.arch}${target.ext}`;
const url = `https://github.com/jackchuka/mdschema/releases/download/v${version}/${archiveName}`;
const binDir = path.join(__dirname, "bin");
const binaryPath = path.join(binDir, `mdschema${ext}`);

function fetch(url, redirectsLeft = 5) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          if (redirectsLeft <= 0) {
            reject(new Error("Too many redirects"));
            return;
          }
          fetch(res.headers.location, redirectsLeft - 1).then(resolve, reject);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: HTTP ${res.statusCode} for ${url}`));
          return;
        }
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

function extractTarGz(buffer) {
  fs.mkdirSync(binDir, { recursive: true });
  const tmpFile = path.join(binDir, "download.tar.gz");
  fs.writeFileSync(tmpFile, buffer);
  // GoReleaser archives place the binary at the root (no subdirectory wrapper)
  execSync(`tar xzf "${tmpFile}" -C "${binDir}" mdschema`, { stdio: "pipe" });
  fs.unlinkSync(tmpFile);
}

function extractZip(buffer) {
  fs.mkdirSync(binDir, { recursive: true });
  const tmpFile = path.join(binDir, "download.zip");
  fs.writeFileSync(tmpFile, buffer);
  execSync(`powershell -command "Expand-Archive -Force '${tmpFile}' '${binDir}'"`, {
    stdio: "pipe",
  });
  fs.unlinkSync(tmpFile);
}

async function main() {
  console.log(`Downloading mdschema v${version} for ${target.os}/${target.arch}...`);

  const buffer = await fetch(url);

  if (target.ext === ".zip") {
    extractZip(buffer);
  } else {
    extractTarGz(buffer);
  }

  if (process.platform !== "win32") {
    fs.chmodSync(binaryPath, 0o755);
  }

  console.log(`mdschema v${version} installed successfully.`);
}

main().catch((err) => {
  console.error(`Failed to install mdschema: ${err.message}`);
  console.error("Try installing manually: https://github.com/jackchuka/mdschema/releases");
  process.exit(1);
});
