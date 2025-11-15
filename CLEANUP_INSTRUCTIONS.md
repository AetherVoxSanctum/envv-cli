# envv Repository Security Cleanup Instructions

This repository contains scripts to remove sensitive business data before making the project public as open source.

## 📋 What Gets Removed

### Files Deleted Completely (from history):
- `BACKEND_DESIGN_PARTNERS.md` - Business strategy and competitive analysis
- `DESIGN_PARTNER_BACKEND_PLAN.md` - B2B SaaS planning documents
- `DESIGN_PARTNER_READY.md` - Design partner checklist
- `DEPLOYMENT_PLAN.md` - Internal deployment strategy

### Files Sanitized:
- `test-demo.sh` - Removes `/Users/wdr/` hardcoded paths
- `demo/setup-working-demo.sh` - Replaces Stripe demo keys with obvious examples
- `demo/.env.example` - Replaces Stripe demo keys with obvious examples

## 🚀 Quick Start

### 1. Run the Cleanup Script

```bash
# From the repository root
./cleanup-sensitive-data.sh
```

**What it does:**
- Creates a backup of your entire repository
- Removes business strategy files from git history
- Sanitizes hardcoded paths and API keys
- Commits the changes
- Provides next steps

**Time:** ~2-5 minutes depending on repository size

### 2. Verify the Cleanup

```bash
# Verify all sensitive data is gone
./verify-cleanup.sh
```

**What it checks:**
- Working directory is clean
- Files removed from git history
- No hardcoded personal paths remain
- Demo keys are sanitized
- Repository size (should be smaller)

### 3. Push to Remote

```bash
# Push the rewritten history
git push --force-with-lease origin claude/audit-sensitive-data-019nJrCLK7BNzmdHL7hRNbp2

# Or push all branches
git push --force-with-lease --all
```

**⚠️ WARNING:** This rewrites remote history! All collaborators must re-clone.

## 📖 Detailed Process

### Before Running the Script

1. **Commit all changes**
   ```bash
   git add -A
   git commit -m "Prepare for cleanup"
   ```

2. **Ensure you're on the correct branch**
   ```bash
   git branch --show-current
   ```

3. **Optional: Install git-filter-repo for better performance**
   ```bash
   # macOS
   brew install git-filter-repo

   # Linux/macOS with Python
   pip install git-filter-repo

   # Debian/Ubuntu
   sudo apt install git-filter-repo
   ```

### Running the Cleanup

The script will:

1. ✅ **Check prerequisites**
   - Verify you're in a git repository
   - Warn about uncommitted changes
   - List what will be removed

2. ✅ **Request confirmation**
   - You must type "YES" to proceed
   - Creates a backup before any changes

3. ✅ **Remove files from history**
   - Uses `git-filter-repo` (preferred) or `git filter-branch`
   - Completely erases files from all commits
   - Recalculates all commit hashes

4. ✅ **Sanitize remaining files**
   - Replaces hardcoded paths
   - Updates demo API keys
   - Makes changes obvious as examples

5. ✅ **Commit sanitized changes**
   - Creates a new commit with cleaned files
   - Includes detailed commit message

### After Running the Script

1. **Review changes**
   ```bash
   # Check recent commits
   git log --oneline -10

   # Verify files are gone from history
   git log --all --full-history -- BACKEND_DESIGN_PARTNERS.md
   # Should return nothing
   ```

2. **Run verification**
   ```bash
   ./verify-cleanup.sh
   ```

3. **Test the repository**
   ```bash
   # Make sure the project still builds
   make install

   # Run tests if available
   make test
   ```

4. **Push to remote**
   ```bash
   # Force push (rewrites remote history)
   git push --force-with-lease
   ```

5. **Update collaborators**
   - Notify all team members
   - They must re-clone or reset:
     ```bash
     git fetch origin
     git reset --hard origin/main
     ```

## 🔍 Verification Checklist

Run `./verify-cleanup.sh` to automatically check:

- [ ] Business strategy files removed from working directory
- [ ] Business strategy files removed from git history
- [ ] No `/Users/wdr/` paths in codebase
- [ ] Demo API keys sanitized with obvious examples
- [ ] Repository size reduced (history cleaned)
- [ ] Project still builds successfully

## 🛟 Troubleshooting

### Script fails with "git-filter-repo not found"
**Solution:** Script will automatically fall back to `git filter-branch`. For better performance, install git-filter-repo:
```bash
pip install git-filter-repo
```

### "You have uncommitted changes"
**Solution:** Commit or stash your changes:
```bash
git stash
./cleanup-sensitive-data.sh
git stash pop
```

### Force push is rejected
**Solution:** Check if branch is protected. You may need to:
1. Temporarily disable branch protection on GitHub
2. Use `--force` instead of `--force-with-lease` (less safe)
3. Contact repository admin

### Repository seems broken after cleanup
**Solution:** Restore from backup:
```bash
# Find your backup
ls -la ../envv-cli-backup-*

# Restore
cd ..
rm -rf envv-cli
cp -r envv-cli-backup-YYYYMMDD-HHMMSS envv-cli
cd envv-cli
```

### Files still show in git history
**Solution:** The cleanup may not have completed. Try:
```bash
# Manually remove from history
git filter-branch --force --index-filter \
  "git rm --cached --ignore-unmatch BACKEND_DESIGN_PARTNERS.md" \
  --prune-empty --tag-name-filter cat -- --all

# Clean up refs
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

## 📊 What Remains (Safe for Open Source)

After cleanup, your repository will contain:

✅ **Core codebase** - All Mozilla SOPS functionality
✅ **Demo application** - Working examples with sanitized keys
✅ **Documentation** - README, usage guides, etc.
✅ **Test files** - All test suites with mock data
✅ **Build scripts** - Makefile, installation scripts
✅ **Example configs** - With obviously fake values

❌ **Business strategy** - Removed
❌ **Personal paths** - Sanitized
❌ **Real-looking keys** - Replaced with obvious examples

## 🎯 Final Steps Before Making Public

1. **Run cleanup script** ✓
2. **Run verification script** ✓
3. **Test build and demos** ✓
4. **Push cleaned history** ✓
5. **Update README** if needed
6. **Make repository public** on GitHub/GitLab
7. **Delete backup** after confirming everything works
8. **Add SECURITY.md** for vulnerability reporting
9. **Add CODE_OF_CONDUCT.md** for community guidelines
10. **Add CONTRIBUTING.md** for contributor guidelines

## 📝 Notes

- **Backup location:** `../envv-cli-backup-YYYYMMDD-HHMMSS/`
- **Keep backup** for at least 1 week after cleanup
- **Git history size:** Should be noticeably smaller after cleanup
- **Commit hashes:** Will all change - this is expected
- **Tags:** Will be preserved and rewritten if they exist

## 🆘 Need Help?

If you encounter issues:

1. Check the backup exists: `ls -la ../envv-cli-backup-*`
2. Review script output for errors
3. Run verification script for details: `./verify-cleanup.sh`
4. Restore from backup if needed
5. Try manual cleanup (see Troubleshooting)

## ⚠️ Important Reminders

- ✅ This rewrites git history - irreversible once pushed!
- ✅ All collaborators must re-clone after force push
- ✅ Protected branches may block force push
- ✅ Keep backup until verified working
- ✅ Test thoroughly before making repository public

---

**Ready to clean?** Run `./cleanup-sensitive-data.sh` to begin!
