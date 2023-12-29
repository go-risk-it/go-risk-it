"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const ws_1 = __importDefault(require("ws"));
const ws = new ws_1.default('http://localhost:8080/ws');
ws.on('open', () => {
    console.log('Connected to server');
    ws.send('Hello, server!');
});
ws.on('message', (message) => {
    console.log(`Received message from server: ${message}`);
});
ws.on('close', () => {
    console.log('Disconnected from server');
});
