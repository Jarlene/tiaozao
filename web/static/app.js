const API_BASE = '/api';

// State
let categories = [];
let uploadedImages = [];

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
      card.innerHTML = `
        <img src="${img}" alt="${product.title}" onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22260%22 height=%22200%22><rect fill=%22%23eee%22 width=%22260%22 height=%22200%22/><text x=%2250%%22 y=%2250%%22 text-anchor=%22middle%22 dy=%22.3em%22 fill=%22%23aaa%22>无图片</text></svg>'">
        <div class="product-info">
          <h3>${escapeHtml(product.title)}</h3>
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

  try {
    const product = await api('/products/' + id);
    const el = document.getElementById('productDetail');
    const images = (product.images || []).map(img => `<img src="${img.url}" alt="">`).join('');

    el.innerHTML = `
      ${images ? `<div class="images">${images}</div>` : ''}
      <h1>${escapeHtml(product.title)}</h1>
      <div class="price">¥${product.price.toFixed(2)}</div>
      ${product.original_price ? `<div class="meta">原价: ¥${product.original_price.toFixed(2)}</div>` : ''}
      <div class="meta">
        ${product.category ? `分类: ${escapeHtml(product.category.name)} | ` : ''}
        状态: ${statusLabel(product.status)} | 发布时间: ${new Date(product.created_at).toLocaleDateString('zh-CN')}
      </div>
      ${product.description ? `<div class="description">${escapeHtml(product.description)}</div>` : ''}
    `;
  } catch (e) {
    document.getElementById('productDetail').innerHTML = `<p>加载失败: ${e.message}</p>`;
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

function statusLabel(s) {
  const map = { active: '在售', draft: '草稿', inactive: '已下架' };
  return map[s] || s;
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
