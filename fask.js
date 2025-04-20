const express = require('express');
const app = express();
const port = 8080;

app.get('/hello', (req, res) => {
  res.send('Hello from Node!');
});

app.listen(port, () => {
  console.log(`Node server running at http://localhost:${port}/hello`);
});
