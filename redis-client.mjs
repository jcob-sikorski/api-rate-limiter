import redis from 'redis';

// Function to create a new Redis client
const createRedisClient = async () => {
    const client = redis.createClient(6379);
    await client.connect();

    // Handle client errors
    client.on('error', (err) => {
        console.error('Redis client error:', err);
    });

    console.log('connected to redis');
    return client;
};

// Connect to the local Redis server
const client = await createRedisClient();


// Function to increment calls in a minute and check the limit
const incrementAndCheckLimit = async (key) => {
    try {
        const currentCalls = await client.get(key) || 0;

        if (currentCalls < 60) {
            await client.set(key, parseInt(currentCalls) + 1, 'EX', 60);
            return true; // Within the limit
        } else {
            return false; // Exceeded the limit
        }
    } catch (error) {
        console.error('Error in incrementAndCheckLimit:', error);
        return false; // Handle the error appropriately
    }
};

// Function to close the Redis client
const closeRedisClient = () => {
    client.quit();
};

export {
    incrementAndCheckLimit,
    closeRedisClient,
};
