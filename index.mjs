// index.mjs

import express from 'express';
import path from 'path';
import { incrementAndCheckLimit } from './redis-client.mjs';

const app = express();
app.use(express.json());

// Use import.meta.url to get the current module's directory
const __filename = new URL(import.meta.url).pathname;
const __dirname = path.dirname(__filename);

app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'index.html'));
});

async function api(req, res) {
    return res.json({ response: 'ok', callsInAMinute: {} });
}

function rateLimiter(req, res, next) {
    const uid = req.params.uid;

    incrementAndCheckLimit(uid)
        .then((withinLimit) => {
            if (withinLimit) {
                next();
            } else {
                return res.status(429).json({ response: 'error', message: 'Rate limit exceeded' });
            }
        })
        .catch((error) => {
            console.error('Error in rateLimiter:', error);
            return res.status(500).json({ response: 'error', message: 'Internal Server Error' });
        });
}

app.post('/api/:uid', rateLimiter, api); // Fix: Specify the uid parameter in the route

app.listen(3000, () => {
    console.log('Server is running on port 3000');
});
