#!/usr/bin/env node
// battlesnake_server.js
// Pure Node.js implementation, no extensions or external libraries

const http = require("http");
const fs = require("fs");
const path = require("path");



// ---- Utility: Dynamic import of a JS module ----
function importDynamic(modulePath) {
    const absPath = path.resolve(modulePath);
    delete require.cache[absPath]; // ensure reload
    return require(absPath);
}

// ---- Utility: Read JSON body ----
function readJson(req) {
    return new Promise((resolve, reject) => {
        let data = "";
        req.on("data", chunk => (data += chunk));
        req.on("end", () => {
            try {
                resolve(data ? JSON.parse(data) : {});
            } catch (err) {
                reject(err);
            }
        });
    });
}

// ---- Main server function ----
function runServer(handlers, port) {
    const host = "0.0.0.0";

    const server = http.createServer(async (req, res) => {
        try {
            res.setHeader("Server", "battlesnake/github/starter-snake-node");

            if (req.method === "GET" && req.url === "/") {
                const info = handlers.info();
                info.author = "IFSP";
                info.apiversion = "1";
                res.writeHead(200, { "Content-Type": "application/json" });
                return res.end(JSON.stringify(info));
            }

            if (req.method === "POST" && req.url === "/start") {
                const gameState = await readJson(req);
                handlers.start(gameState);
                res.writeHead(200, { "Content-Type": "text/plain" });
                return res.end("ok");
            }

            if (req.method === "POST" && req.url === "/move") {
                const gameState = await readJson(req);
                const move = handlers.move(gameState);
                res.writeHead(200, { "Content-Type": "application/json" });
                return res.end(JSON.stringify(move));
            }

            if (req.method === "POST" && req.url === "/end") {
                const gameState = await readJson(req);
                handlers.end(gameState);
                res.writeHead(200, { "Content-Type": "text/plain" });
                return res.end("ok");
            }

            // Not found
            res.writeHead(404, { "Content-Type": "text/plain" });
            res.end("Not found");
        } catch (err) {
            console.error("Error handling request:", err);
            res.writeHead(500, { "Content-Type": "text/plain" });
            res.end("Internal server error");
        }
    });

    server.listen(port, host, () => {
        console.log(`\nRunning Battlesnake at http://${host}:${port}`);
    });
}

// ---- Entry point ----
if (require.main === module) {
    const [, , snakePath, portArg] = process.argv;
    if (!snakePath || !portArg) {
        console.error("Usage: node battlesnake_server.js <snake.js> <port>");
        process.exit(1);
    }

    const port = parseInt(portArg, 10);
    const snake = importDynamic(snakePath);

    runServer(
        {
            info: snake.info,
            start: snake.start,
            move: snake.move,
            end: snake.end,
        },
        port
    );
}
