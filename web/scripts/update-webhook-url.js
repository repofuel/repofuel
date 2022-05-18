const {createAppAuth} = require('@octokit/auth-app');
const fetch = require('node-fetch');
const fs = require('fs');
const path = require('path');
const dotenv = require('dotenv');

dotenv.config({path: path.resolve(process.cwd(), 'local.env')});

async function main() {
  const res = await fetch('http://127.0.0.1:4040/api/tunnels');
  const resBody = await res.json();
  console.log('ngrok host:', resBody.tunnels[0].public_url);

  const privateKeyPath = path.join(
    __dirname,
    '..',
    process.env.PRIVATE_KEY_PATH
  );

  const authOptions = {
    appId: process.env.GITHUB_APP_ID,
    privateKey: fs.readFileSync(privateKeyPath, 'utf8'),
  };

  const auth = createAppAuth(authOptions);
  const {token} = await auth({type: 'app'});

  await fetch('https://api.github.com/app/hook/config', {
    method: 'PATCH',
    body: JSON.stringify({
      url: resBody.tunnels[0].public_url + '/ingest/apps/github/webhook',
    }),
    headers: {
      accept: 'application/vnd.github.v3+json',
      Authorization: 'Bearer ' + token,
    },
  })
    .catch(console.log)
    .then(() => console.log('Done!'));
}

main();
