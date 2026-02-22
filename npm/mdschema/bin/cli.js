#!/usr/bin/env node

"use strict";

const { spawnSync } = require("child_process");
const { getBinaryPath } = require("../lib/platform");

const binary = getBinaryPath();
const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  env: process.env,
});

if (result.error) {
  if (result.error.code === "ENOENT") {
    console.error(
      `mdschema binary not found at ${binary}.\n` +
        "Try reinstalling: npm install @jackchuka/mdschema"
    );
  } else {
    console.error(`Failed to run mdschema: ${result.error.message}`);
  }
  process.exit(1);
}

process.exit(result.status ?? 1);
