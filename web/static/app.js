const API_BASE = '/api';

// State
let categories = [];
let uploadedImages = [];
let currentProductId = null;

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
  await loadCategories();
  const page = window.location.pathname;
  if (page === '/publish.html') initPublishPage();
  else if (page.startsWith('/detail.html')) initDetailPage();
  else initHomePage();
});

async function api(path, options = {}) {
  const resp = await fetch(API_BASE + path, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });
  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: resp.statusText }));
    throw new Error(err.error || 'Request failed');
  }
  return resp.json();
}

async function loadCategories() {
  try {
    categories = await api('/categories');
    return categories;
  } catch (e) {
    console.error('Failed to load categories:', e);
    return [];
  }
}

// Home Page
function initHomePage() {
  const filters = document.querySelector('.filters');
  if (filters) {
    const sel = document.getElementById('categoryFilter');
    if (sel) renderCategoryOptions(sel, categories);
  }
  loadProducts();
}

async function loadProducts(page = 1) {
  const grid = document.getElementById('productGrid');
  if (!grid) return;
  grid.innerHTML = '<p>加载中...</p>';

  try {
    const categoryId = document.getElementById('categoryFilter')?.value || '';
    const params = new URLSearchParams({ page, size: 12 });
    if (categoryId) params.set('category_id', categoryId);

    const data = await api('/products?' + params.toString());
    grid.innerHTML = '';

    if (data.items.length === 0) {
      grid.innerHTML = '<p class="empty">暂无商品</p>';
      return;
    }

    data.items.forEach(product => {
      const card = document.createElement('div');
      card.className = 'product-card';
      card.onclick = () => window.location = '/detail.html?id=' + product.id;

      const img = product.images?.[0]?.url || '';
      const ratingHtml = product.rating && product.rating.count > 0
        ? `<div class="rating-summary"><span class="stars">${'★'.repeat(Math.round(product.rating.average))}${'☆'.repeat(5 - Math.round(product.rating.average))}</span> ${product.rating.average.toFixed(1)} (${product.rating.count})</div>`
        : '';
      card.innerHTML = `
        <img src="${img}" alt="${product.title}" onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22260%22 height=%22200%22><rect fill=%22%23eee%22 width=%22260%22 height=%22200%22/><text x=%2250%%22 y=%2250%%22 text-anchor=%22middle%22 dy=%22.3em%22 fill=%22%23aaa%22>无图片</text></svg>'">
        <div class="product-info">
          <h3>${escapeHtml(product.title)}</h3>
          ${ratingHtml}
          <div>
            <span class="product-price">¥${product.price.toFixed(2)}</span>
            ${product.original_price ? `<span class="product-original-price">¥${product.original_price.toFixed(2)}</span>` : ''}
          </div>
        </div>
      `;
      grid.appendChild(card);
    });

    renderPagination(data.total, data.page, data.size);
  } catch (e) {
    grid.innerHTML = `<p class="empty">加载失败: ${e.message}</p>`;
  }
}

function renderPagination(total, page, size) {
  const el = document.getElementById('pagination');
  if (!el) return;
  const totalPages = Math.ceil(total / size);
  if (totalPages <= 1) { el.innerHTML = ''; return; }

  let html = '';
  for (let i = 1; i <= totalPages; i++) {
    html += `<button class="${i === page ? 'active' : ''}" onclick="loadProducts(${i})">${i}</button>`;
  }
  el.innerHTML = html;
}

// Publish Page
function initPublishPage() {
  const catSelect = document.getElementById('category');
  if (catSelect) renderCategoryOptions(catSelect, categories);

  const dropZone = document.getElementById('dropZone');
  const fileInput = document.getElementById('fileInput');
  if (dropZone && fileInput) {
    dropZone.onclick = () => fileInput.click();
    dropZone.ondragover = (e) => { e.preventDefault(); dropZone.style.borderColor = '#e74c3c'; };
    dropZone.ondragleave = () => { dropZone.style.borderColor = '#ddd'; };
    dropZone.ondrop = (e) => { e.preventDefault(); handleFiles(e.dataTransfer.files); };
    fileInput.onchange = () => handleFiles(fileInput.files);
  }

  const form = document.getElementById('publishForm');
  if (form) {
    form.onsubmit = async (e) => {
      e.preventDefault();
      const title = document.getElementById('title').value.trim();
      const price = parseFloat(document.getElementById('price').value);
      const categoryId = document.getElementById('category').value;
      const description = document.getElementById('description').value.trim();
      const originalPrice = parseFloat(document.getElementById('originalPrice').value) || 0;

      if (!title) { alert('请输入商品标题'); return; }
      if (!price || price <= 0) { alert('请输入有效价格'); return; }

      try {
        const body = {
          title,
          description,
          price,
          original_price: originalPrice || undefined,
          category_id: categoryId ? parseInt(categoryId) : null,
          image_ids: uploadedImages.map(i => i.id),
          cover_image_id: uploadedImages[0]?.id || null,
        };
        await api('/products', {
          method: 'POST',
          body: JSON.stringify(body),
        });
        alert('发布成功!');
        window.location = '/';
      } catch (e) {
        alert('发布失败: ' + e.message);
      }
    };
  }
}

async function handleFiles(files) {
  for (const file of files) {
    if (!file.type.startsWith('image/')) continue;
    if (file.size > 10 * 1024 * 1024) { alert('图片最大 10MB'); continue; }

    const formData = new FormData();
    formData.append('file', file);

    try {
      const resp = await fetch(API_BASE + '/products/images', {
        method: 'POST',
        body: formData,
      });
      if (!resp.ok) throw new Error((await resp.json()).error);
      const data = await resp.json();
      uploadedImages.push(data);

      const preview = document.getElementById('imagePreview');
      if (preview) {
        const img = document.createElement('img');
        img.src = data.url;
        preview.appendChild(img);
      }
    } catch (e) {
      alert('上传失败: ' + e.message);
    }
  }
}

// Detail Page
async function initDetailPage() {
  const params = new URLSearchParams(window.location.search);
  const id = params.get('id');
  if (!id) { document.getElementById('productDetail').innerHTML = '<p>缺少商品ID</p>'; return; }
  currentProductId = parseInt(id);

  setupRatingPicker();

  try {
    const product = await api('/products/' + id);
    const el = document.getElementById('productDetail');
    const images = (product.images || []).map(img => `<img src="${img.url}" alt="">`).join('');
    const ratingHtml = product.rating && product.rating.count > 0
      ? `<div class="rating-summary">综合评分：<span class="stars">${'★'.repeat(Math.round(product.rating.average))}${'☆'.repeat(5 - Math.round(product.rating.average))}</span> ${product.rating.average.toFixed(1)} / 5（${product.rating.count} 条评价）</div>`
      : `<div class="rating-summary">暂无评价</div>`;

    el.innerHTML = `
      ${images ? `<div class="images">${images}</div>` : ''}
      <h1>${escapeHtml(product.title)}</h1>
      ${ratingHtml}
      <div class="price">¥${product.price.toFixed(2)}</div>
      ${product.original_price ? `<div class="meta">原价: ¥${product.original_price.toFixed(2)}</div>` : ''}
      <div class="meta">
        ${product.category ? `分类: ${escapeHtml(product.category.name)} | ` : ''}
        库存: ${product.stock ?? 0} | 销量: ${product.sales ?? 0} | 状态: ${statusLabel(product.status)} | 发布时间: ${new Date(product.created_at).toLocaleDateString('zh-CN')}
      </div>
      ${product.description ? `<div class="description">${escapeHtml(product.description)}</div>` : ''}
    `;
  } catch (e) {
    document.getElementById('productDetail').innerHTML = `<p>加载失败: ${e.message}</p>`;
  }

  loadComments();
}

async function loadComments() {
  const list = document.getElementById('commentList');
  if (!list || currentProductId == null) return;
  list.innerHTML = '加载中...';
  try {
    const data = await api(`/products/${currentProductId}/comments?page=1&page_size=10`);
    renderComments(data);
  } catch (e) {
    list.innerHTML = `<div class="comments-empty">评论加载失败: ${e.message}</div>`;
  }
}

function renderComments(data) {
  const list = document.getElementById('commentList');
  const comments = data.comments || [];
  if (comments.length === 0) {
    list.innerHTML = '<div class="comments-empty">还没有评价，来抢沙发吧～</div>';
    return;
  }

  list.innerHTML = comments.map(c => renderCommentItem(c)).join('') +
    (data.total > comments.length ? `<div class="comments-empty">共 ${data.total} 条评价</div>` : '');
}

function renderCommentItem(c) {
  const stars = c.rating
    ? `<span class="stars">${'★'.repeat(c.rating)}${'☆'.repeat(5 - c.rating)}</span>`
    : '';
  const replies = (c.replies || []).map(r => `
    <div class="comment-item comment-reply">
      <div class="comment-header">
        <span class="comment-user">${escapeHtml(r.user_nick || '用户' + r.user_id)}</span>
        <span class="comment-time">${formatTime(r.created_at)}</span>
      </div>
      <div class="comment-content">${escapeHtml(r.content)}</div>
    </div>
  `).join('');
  return `
    <div class="comment-item" id="comment-${c.id}">
      <div class="comment-header">
        <span class="comment-user">${escapeHtml(c.user_nick || '用户' + c.user_id)}</span>
        <span class="comment-time">${formatTime(c.created_at)}</span>
      </div>
      <div class="comment-content">${stars ? stars + ' ' : ''}${escapeHtml(c.content)}</div>
      <button type="button" class="reply-toggle" onclick="toggleReply(${c.id})">回复</button>
      <div class="comment-replies">${replies}</div>
      <div class="reply-form" id="replyForm-${c.id}" style="display:none;">
        <div style="display:flex; gap:8px;">
          <input type="text" id="replyInput-${c.id}" placeholder="回复这条评价...">
          <button type="button" class="btn-primary" onclick="submitReply(${c.id})">提交</button>
        </div>
      </div>
    </div>
  `;
}

function toggleReply(commentId) {
  const form = document.getElementById('replyForm-' + commentId);
  if (form) form.style.display = form.style.display === 'none' ? 'flex' : 'none';
}

async function submitReply(commentId) {
  const input = document.getElementById('replyInput-' + commentId);
  const content = input.value.trim();
  if (!content) { alert('请输入回复内容'); return; }
  try {
    await api(`/products/${currentProductId}/comments/${commentId}/reply`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    });
    input.value = '';
    loadComments();
  } catch (e) {
    alert('回复失败: ' + e.message);
  }
}

function setupRatingPicker() {
  const box = document.getElementById('ratingStars');
  if (!box) return;
  box.onmouseover = (e) => {
    const v = e.target.dataset.value;
    if (v) paintStars(box, parseInt(v));
  };
  box.onmouseleave = () => paintStars(box, parseInt(box.dataset.rating || '0', 10));
  box.onclick = (e) => {
    const v = e.target.dataset.value;
    if (v) {
      box.dataset.rating = v;
      paintStars(box, parseInt(v));
    }
  };
}

function paintStars(box, n) {
  [...box.children].forEach((span, i) => {
    span.style.color = i < n ? '#f5a623' : '#ddd';
  });
}

async function submitComment() {
  const content = document.getElementById('commentContent').value.trim();
  const rating = parseInt(document.getElementById('ratingStars').dataset.rating || '0', 10);
  if (!content) { alert('请输入评价内容'); return; }
  try {
    await api(`/products/${currentProductId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content, rating }),
    });
    document.getElementById('commentContent').value = '';
    const stars = document.getElementById('ratingStars');
    stars.dataset.rating = '0';
    paintStars(stars, 0);
    loadComments();
  } catch (e) {
    alert('发表失败: ' + e.message);
  }
}

// Utilities
function renderCategoryOptions(select, cats) {
  select.innerHTML = '<option value="">不限</option>';
  cats.forEach(cat => {
    select.innerHTML += `<option value="${cat.id}">${escapeHtml(cat.name)}</option>`;
    if (cat.children) {
      cat.children.forEach(child => {
        select.innerHTML += `<option value="${child.id}">&nbsp;&nbsp;${escapeHtml(child.name)}</option>`;
      });
    }
  });
}

// Product status int enum: 1=active, 2=sold, 3=offline
function statusLabel(s) {
  const map = { 1: '在售', 2: '已售出', 3: '已下架' };
  return map[s] || ('未知(' + s + ')');
}

function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString('zh-CN');
  } catch { return iso; }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
