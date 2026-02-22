"use strict";

const path = require("path");

const PLATFORMS = {
  darwin: {
    arm64: "@jackchuka/mdschema-darwin-arm64",
    x64: "@jackchuka/mdschema-darwin-x64",
  },
  linux: {
    arm64: "@jackchuka/mdschema-linux-arm64",
    x64: "@jackchuka/mdschema-linux-x64",
  },
  win32: {
    arm64: "@jackchuka/mdschema-windows-arm64",
    x64: "@jackchuka/mdschema-windows-x64",
  },
};

function getPackageName() {
  const platform = process.platform;
  const arch = process.arch;

  const archMap = PLATFORMS[platform];
  if (!archMap) {
    throw new Error(
      `Unsupported platform: ${platform}. mdschema supports darwin, linux, and win32.`
    );
  }

  const pkg = archMap[arch];
  if (!pkg) {
    throw new Error(
      `Unsupported architecture: ${arch} on ${platform}. mdschema supports arm64 and x64.`
    );
  }

  return pkg;
}

function getBinaryPath() {
  const ext = process.platform === "win32" ? ".exe" : "";
  const pkg = getPackageName();

  // Try to resolve from optionalDependencies first
  try {
    return require.resolve(`${pkg}/bin/mdschema${ext}`);
  } catch {
    // Fall back to locally downloaded binary (from postinstall)
    return path.join(__dirname, "..", "bin", `mdschema${ext}`);
  }
}

// Map Node.js platform/arch to GoReleaser naming for download URLs
function getGoReleaserTarget() {
  const platform = process.platform;
  const arch = process.arch;

  const osMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { arm64: "arm64", x64: "amd64" };

  const os = osMap[platform];
  const goarch = archMap[arch];

  if (!os || !goarch) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }

  return { os, arch: goarch, ext: platform === "win32" ? ".zip" : ".tar.gz" };
}

module.exports = { getBinaryPath, getPackageName, getGoReleaserTarget };
