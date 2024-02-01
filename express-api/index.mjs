import express from 'express';
import path from 'path';
import amqp from 'amqplib/callback_api.js';

const app = express();
app.use(express.json());

// Use import.meta.url to get the current module's directory
const __filename = new URL(import.meta.url).pathname;
const __dirname = path.dirname(__filename);

function connectWithRetry() {
    console.log('Connecting to RabbitMQ...');
    try {
        amqp.connect('amqp://guest:guest@rabbitmq:5672/', function(error0, connection) {
            if (error0) {
                console.error('Failed to connect to RabbitMQ:', error0);
                console.log('Retrying in 5 seconds...');
                setTimeout(connectWithRetry, 5000);
                return;
            }

            // Connection successful, now we can create a channel
            connection.createChannel(function(error1, channel) {
                if (error1) {
                    throw error1;
                }

                var queue = 'direct_queue';

                // Assert a queue into existence. This operation is idempotent.
                channel.assertQueue(queue, {
                    durable: false
                });

                console.log(" [*] Waiting for messages in %s. To exit press CTRL+C", queue);

                // Consume messages from the queue
                channel.consume(queue, function(msg) {
                    console.log(" [x] Received %s", msg.content.toString());
                }, {
                    noAck: true
                });
            });
        });
    } catch (error) {
        console.error('An error occurred:', error);
    }
}

// Start the connection process
connectWithRetry();

app.listen(3000, () => {
    console.log('Server is running on port 3000');
});