Three components: API, queue worker and web client. The API never talks
to the worker directly; everything goes through the queue.
