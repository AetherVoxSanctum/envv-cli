let currentView = 'list';
let posts = [];

async function init() {
    await checkConfig();
    await loadPosts();
    setupEventListeners();
}

async function checkConfig() {
    try {
        const response = await fetch('/api/config');
        const config = await response.json();

        const analyticsStatus = document.getElementById('analyticsStatus');
        const paymentsStatus = document.getElementById('paymentsStatus');

        if (config.analyticsEnabled) {
            analyticsStatus.textContent = 'Active';
            analyticsStatus.className = 'status-value active';
            console.log('Analytics configured:', config.googleAnalyticsId, config.mixpanelToken);
        } else {
            analyticsStatus.textContent = 'Missing';
            analyticsStatus.className = 'status-value inactive';
        }

        if (config.paymentsEnabled) {
            paymentsStatus.textContent = 'Active';
            paymentsStatus.className = 'status-value active';
        } else {
            paymentsStatus.textContent = 'Missing';
            paymentsStatus.className = 'status-value inactive';
        }
    } catch (error) {
        console.error('Failed to check config:', error);
    }
}

async function loadPosts() {
    try {
        const response = await fetch('/api/posts');
        posts = await response.json();
        renderPostsList();
    } catch (error) {
        console.error('Failed to load posts:', error);
        document.getElementById('postsList').innerHTML =
            '<div class="error">Failed to load posts. Please try again later.</div>';
    }
}

function renderPostsList() {
    const container = document.getElementById('postsList');

    if (posts.length === 0) {
        container.innerHTML = '<div class="loading">No posts found</div>';
        return;
    }

    container.innerHTML = posts.map(post => `
        <div class="post-card" onclick="showPost('${post.slug}')">
            <h3>${post.title}</h3>
            <div class="post-meta">
                By ${post.author} • ${new Date(post.date).toLocaleDateString()}
            </div>
            <p class="post-excerpt">${post.excerpt}</p>
            <div class="post-tags">
                ${post.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
            </div>
        </div>
    `).join('');
}

async function showPost(slug) {
    try {
        const response = await fetch(`/api/posts/${slug}`);
        const post = await response.json();

        document.getElementById('posts').classList.add('hidden');
        document.getElementById('postDetail').classList.remove('hidden');

        document.getElementById('postContent').innerHTML = `
            <h1>${post.title}</h1>
            <div class="post-meta">
                By ${post.author} • ${new Date(post.date).toLocaleDateString()}
            </div>
            <div class="post-content">${post.content}</div>
        `;

        currentView = 'detail';
        window.scrollTo(0, 0);
    } catch (error) {
        console.error('Failed to load post:', error);
    }
}

function showPostsList() {
    document.getElementById('postDetail').classList.add('hidden');
    document.getElementById('posts').classList.remove('hidden');
    currentView = 'list';
}

function setupEventListeners() {
    document.getElementById('subscribeForm').addEventListener('submit', async (e) => {
        e.preventDefault();

        const email = document.getElementById('email').value;
        const messageDiv = document.getElementById('subscribeMessage');

        try {
            const response = await fetch('/api/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email })
            });

            const result = await response.json();

            if (result.success) {
                messageDiv.innerHTML = `
                    <div class="message success">
                        ${result.premium ? '✨ Premium subscription successful!' : '✓ Subscription successful!'}
                        ${result.message}
                    </div>
                `;
                document.getElementById('email').value = '';
            } else {
                throw new Error(result.error);
            }
        } catch (error) {
            messageDiv.innerHTML = `
                <div class="message error">
                    ✗ Subscription failed: ${error.message}
                </div>
            `;
        }

        setTimeout(() => {
            messageDiv.innerHTML = '';
        }, 5000);
    });

    document.getElementById('statsLink').addEventListener('click', (e) => {
        e.preventDefault();
        document.getElementById('stats').classList.toggle('hidden');
    });
}

async function fetchStats() {
    const secretKey = document.getElementById('secretKey').value;
    const messageDiv = document.getElementById('statsMessage');
    const contentDiv = document.getElementById('statsContent');

    if (!secretKey) {
        messageDiv.innerHTML = '<div class="message error">Please enter the backend secret key</div>';
        return;
    }

    try {
        const response = await fetch('/api/stats', {
            headers: {
                'Authorization': `Bearer ${secretKey}`
            }
        });

        if (!response.ok) {
            throw new Error('Invalid secret key');
        }

        const stats = await response.json();

        contentDiv.innerHTML = `
            <div class="stat-card">
                <h3>Total Posts</h3>
                <div class="value">${stats.totalPosts}</div>
            </div>
            <div class="stat-card">
                <h3>Subscribers</h3>
                <div class="value">${stats.totalSubscribers}</div>
            </div>
            <div class="stat-card">
                <h3>Monthly Revenue</h3>
                <div class="value">$${stats.monthlyRevenue}</div>
            </div>
            <div class="stat-card">
                <h3>Active Users</h3>
                <div class="value">${stats.activeUsers}</div>
            </div>
        `;

        contentDiv.classList.remove('hidden');
        messageDiv.innerHTML = '<div class="message success">Stats loaded successfully</div>';
    } catch (error) {
        messageDiv.innerHTML = `<div class="message error">Failed to load stats: ${error.message}</div>`;
        contentDiv.classList.add('hidden');
    }

    setTimeout(() => {
        messageDiv.innerHTML = '';
    }, 5000);
}

document.addEventListener('DOMContentLoaded', init);