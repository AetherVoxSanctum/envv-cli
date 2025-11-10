const express = require('express');
const path = require('path');
const fs = require('fs').promises;
const { marked } = require('marked');
const matter = require('gray-matter');

const app = express();
const PORT = process.env.PORT || 3000;

const ANALYTICS_KEY_1 = process.env.ANALYTICS_KEY_GOOGLE;
const ANALYTICS_KEY_2 = process.env.ANALYTICS_KEY_MIXPANEL;
const PAYMENT_API_KEY = process.env.STRIPE_API_KEY;
const BACKEND_SECRET = process.env.BACKEND_SECRET_KEY;

app.use(express.static(path.join(__dirname, '../public')));
app.use(express.json());

app.get('/api/config', (req, res) => {
  res.json({
    analyticsEnabled: !!(ANALYTICS_KEY_1 && ANALYTICS_KEY_2),
    paymentsEnabled: !!PAYMENT_API_KEY,
    googleAnalyticsId: ANALYTICS_KEY_1 ? `GA-${ANALYTICS_KEY_1.slice(-8)}` : null,
    mixpanelToken: ANALYTICS_KEY_2 ? `${ANALYTICS_KEY_2.slice(0, 4)}...` : null
  });
});

app.get('/api/posts', async (req, res) => {
  try {
    const postsDir = path.join(__dirname, '../posts');
    const files = await fs.readdir(postsDir);
    const posts = [];

    for (const file of files) {
      if (file.endsWith('.md')) {
        const content = await fs.readFile(path.join(postsDir, file), 'utf-8');
        const { data, content: markdownContent } = matter(content);
        posts.push({
          slug: file.replace('.md', ''),
          title: data.title || 'Untitled',
          date: data.date || new Date().toISOString(),
          author: data.author || 'Anonymous',
          excerpt: data.excerpt || markdownContent.slice(0, 150) + '...',
          tags: data.tags || []
        });
      }
    }

    posts.sort((a, b) => new Date(b.date) - new Date(a.date));
    res.json(posts);
  } catch (error) {
    console.error('Error reading posts:', error);
    res.status(500).json({ error: 'Failed to load posts' });
  }
});

app.get('/api/posts/:slug', async (req, res) => {
  try {
    const filePath = path.join(__dirname, '../posts', `${req.params.slug}.md`);
    const content = await fs.readFile(filePath, 'utf-8');
    const { data, content: markdownContent } = matter(content);
    const html = marked(markdownContent);

    res.json({
      ...data,
      slug: req.params.slug,
      content: html
    });
  } catch (error) {
    console.error('Error reading post:', error);
    res.status(404).json({ error: 'Post not found' });
  }
});

app.post('/api/subscribe', (req, res) => {
  const { email } = req.body;

  if (!email) {
    return res.status(400).json({ error: 'Email required' });
  }

  if (PAYMENT_API_KEY) {
    console.log(`[PAYMENT] Processing subscription for: ${email}`);
    console.log(`[PAYMENT] Using Stripe API Key: ${PAYMENT_API_KEY.slice(0, 7)}...`);
  }

  if (ANALYTICS_KEY_2) {
    console.log(`[ANALYTICS] Tracking subscription event with Mixpanel`);
  }

  res.json({
    success: true,
    message: 'Subscription successful',
    premium: !!PAYMENT_API_KEY
  });
});

app.get('/api/stats', (req, res) => {
  const authHeader = req.headers.authorization;

  if (!authHeader || authHeader !== `Bearer ${BACKEND_SECRET}`) {
    return res.status(401).json({ error: 'Unauthorized' });
  }

  res.json({
    totalPosts: 4,
    totalSubscribers: 127,
    monthlyRevenue: 3420,
    activeUsers: 89
  });
});

app.listen(PORT, () => {
  console.log(`\n🚀 Blog server running on http://localhost:${PORT}`);
  console.log('\n🔐 Secret Status (using envv for encrypted storage):');
  console.log(`  ✓ Google Analytics: ${ANALYTICS_KEY_1 ? '🔒 Encrypted & Loaded' : '❌ Not Available'}`);
  console.log(`  ✓ Mixpanel Analytics: ${ANALYTICS_KEY_2 ? '🔒 Encrypted & Loaded' : '❌ Not Available'}`);
  console.log(`  ✓ Stripe Payments: ${PAYMENT_API_KEY ? '🔒 Encrypted & Loaded' : '❌ Not Available'}`);
  console.log(`  ✓ Backend Secret: ${BACKEND_SECRET ? '🔒 Encrypted & Loaded' : '❌ Not Available'}`);

  if (!ANALYTICS_KEY_1 || !ANALYTICS_KEY_2 || !PAYMENT_API_KEY || !BACKEND_SECRET) {
    console.log('\n⚠️  Some secrets are missing!');
    console.log('   Run "envv exec -- npm start" to load encrypted secrets');
    console.log('   Or "envv inject .env" to create a temporary decrypted file\n');
  } else {
    console.log('\n✅ All secrets successfully loaded from encrypted storage!\n');
  }
});