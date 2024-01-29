import express from 'express';
import path from 'path';

const app = express();
app.use(express.json());

// Use import.meta.url to get the current module's directory
const __filename = new URL(import.meta.url).pathname;
const __dirname = path.dirname(__filename);

app.get('/api/:uid', (req, res) => {
    res.json({ response: 'ok' });
});

app.listen(3000, () => {
    console.log('Server is running on port 3000');
});
