# Load Tests

Run load tests on SideKick with [k6](https://k6.io/docs/).

## Dependencies

- Install [k7](https://k6.io/docs/get-started/installation/)
- Install `npm`

## Running

Setup your env variables in a .env file:

```bash
AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID>
AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY>
AWS_REGION=<AWS_REGION>
BUCKET=<BUCKET>
```

Run sidekick locally in a different terminal and run the following commands:

```bash
npm install
npm run bundle
npm run get
```
