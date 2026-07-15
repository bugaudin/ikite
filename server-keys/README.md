# Server SSH keys

Place deployment SSH private keys here locally. This directory is **gitignored**.

```bash
chmod 600 your-key.pem
ssh -i your-key.pem user@your-server
```

Never commit private keys or add them to the repository.
