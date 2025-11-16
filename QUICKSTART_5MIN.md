# envv - 5 Minute Quick Start (Design Partners)

**Time:** 5 minutes | **Difficulty:** Easy | **Result:** Encrypted secrets working in your app

## 🚀 Minute 1: Install envv

```bash
# Clone and build (takes ~60 seconds)
git clone https://github.com/AetherVoxSanctum/envv.git
cd envv
make install

# Add to PATH
export PATH=$PATH:~/go/bin
```

✅ **Check:** Run `~/go/bin/envv --version` - should show version info

## 🔐 Minute 2: Try the Demo

```bash
# Go to demo directory
cd demo

# Run the interactive setup
chmod +x setup-working-demo.sh
./setup-working-demo.sh
```

✅ **Check:** You should see green checkmarks and "Demo setup complete!"

## 📝 Minute 3: Use Encrypted Secrets

```bash
# See your encrypted secrets (gibberish)
cat .env.encrypted

# List secret names (not values!)
./envv list

# Set a new secret
./envv set MY_API_KEY "super_secret_123"

# Run app with secrets
./envv exec npm start
```

✅ **Check:** Server starts with message "All secrets successfully loaded from encrypted storage!"

## 👥 Minute 4: Understand Team Workflow

**The Problem (Traditional):**
```bash
# Everyone can see secrets 😱
cat .env
STRIPE_KEY=sk_live_abcd1234  # EXPOSED!
```

**The Solution (envv):**
```bash
# Secrets are encrypted 🔐
cat .env.encrypted
{
  "data": "ENC[AES256_GCM,data:9zjgqx...]"  # SAFE!
}

# But app still works!
./envv exec npm start  # Secrets loaded in memory only
```

## ⚡ Minute 5: Your Turn!

Try this in YOUR project:

```bash
# Go to your project
cd ~/your-project

# Copy the working demo setup
cp -r ~/envv/demo/keys .
cp ~/envv/demo/.sops.yaml .
cp ~/envv/demo/envv .

# Encrypt your existing .env
export SOPS_AGE_KEY_FILE=$(pwd)/keys/demo.agekey
~/go/bin/envv -e .env > .env.encrypted

# Remove plaintext!
rm .env

# Run your app
./envv exec npm start  # or your start command
```

---

## 🎯 What You Just Achieved

✅ **Secrets are encrypted at rest** - No plaintext on disk
✅ **App runs normally** - Transparent decryption
✅ **Team can share safely** - Commit .env.encrypted to git
✅ **No secrets in logs/history** - They never exist as files

## 🚨 Real Impact

**Before envv:**
- `git add .` → Accidentally commits secrets → **DATA BREACH**
- Screen share → Shows .env file → **SECRETS LEAKED**
- Laptop stolen → Thief has all secrets → **COMPROMISED**

**With envv:**
- `git add .` → Only encrypted file → **SAFE**
- Screen share → Shows encrypted blob → **SAFE**
- Laptop stolen → Thief gets encrypted data → **SAFE**

## 📊 Design Partner Feedback Needed

We want to know:
1. ⏱️ Did this take 5 minutes?
2. 🐛 Did everything work?
3. 💡 What would make this better?
4. 🎯 Would your team use this?

**Contact:** [your-email] | Slack: [your-slack]

---

## 🆘 Troubleshooting

**"command not found: age"**
```bash
brew install age  # macOS
# or
apt-get install age  # Linux
```

**"envv: command not found"**
```bash
export PATH=$PATH:~/go/bin
# or use full path:
~/go/bin/envv
```

**"failed to decrypt"**
```bash
export SOPS_AGE_KEY_FILE=$(pwd)/keys/demo.agekey
```

**Still stuck?** Reach out immediately - we want this to work for you!

---

## 🎬 Live Demo Script (30 seconds)

Perfect for showing your team:

```bash
# The hook (5 seconds)
echo "Our API keys are encrypted:"
cat .env.encrypted | head -3

# The problem it solves (10 seconds)
echo "Without envv, secrets are everywhere:"
echo "- In .env files"
echo "- In git history"
echo "- In bash history"
echo "- On every developer's laptop"

# The solution (10 seconds)
echo "With envv, secrets are always encrypted:"
./envv list  # Shows keys, not values
./envv exec npm start  # App works perfectly

# The closer (5 seconds)
echo "Setup time: 5 minutes"
echo "Security improvement: Priceless"
```

---

**Ready to secure your secrets? You just did! 🎉**