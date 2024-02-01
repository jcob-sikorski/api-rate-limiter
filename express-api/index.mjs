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

                // **   competing consumers pattern   **

                // don’t dispatch a new message to a worker until 
                // it has processed and acknowledged the previous one
                channel.prefetch(1);

                // Consume messages from the queue
                channel.consume(queue, function(msg) {
                    console.log(" [x] Received %s", msg.content.toString());
                
                    // simulate processing the task for random period of time
                    let processing_time = Math.random() * 0.01;
                    console.log(`Received: ${msg.content.toString()}, will take ${processing_time} seconds to process`);
                    setTimeout(() => {
                        channel.ack(msg);
                        console.log('Message processed');
                    }, processing_time * 1000);
                }, {
                    noAck: false
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