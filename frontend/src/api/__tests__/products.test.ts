import { describe, it, expect, vi, beforeEach } from 'vitest'

// Use vi.hoisted to define mocks before vi.mock factory calls
const { mockGet, mockPost, mockPut, mockDelete } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPut: vi.fn(),
  mockDelete: vi.fn(),
}))

// Must define the mock client WITHIN the factory since it's hoisted
vi.mock('../client', () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: mockPut,
    delete: mockDelete,
  },
  ApiResponse: class {},
}))

import { productAPI } from '../products'

describe('productAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch product list with pagination', async () => {
      const mockData = {
        items: [{ id: 1, title: '商品1', price: 1000, status: 1, user_id: 1, created_at: '2024-01-01', images: [] }],
        total: 1, page: 1, page_size: 20, total_pages: 1,
      }
      mockGet.mockResolvedValue({ data: { data: mockData } })

      const result = await productAPI.list(1, 20)
      expect(mockGet).toHaveBeenCalledWith('/products', { params: { page: 1, size: 20 } })
      expect(result).toEqual(mockData)
    })

    it('should use default page and size', async () => {
      mockGet.mockResolvedValue({ data: { data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } })

      await productAPI.list()
      expect(mockGet).toHaveBeenCalledWith('/products', { params: { page: 1, size: 20 } })
    })
  })

  describe('search', () => {
    it('should search with keyword', async () => {
      mockGet.mockResolvedValue({ data: { data: { items: [{ id: 1, title: '手机' }], total: 1, page: 1, page_size: 20, total_pages: 1 } } })

      const result = await productAPI.search({ keyword: '手机' })
      expect(mockGet).toHaveBeenCalledWith('/products/search', { params: { keyword: '手机' } })
      expect(result.items[0].title).toBe('手机')
    })

    it('should search with price range', async () => {
      mockGet.mockResolvedValue({ data: { data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } })

      await productAPI.search({ price_min: 1000, price_max: 5000 })
      expect(mockGet).toHaveBeenCalledWith('/products/search', { params: { price_min: 1000, price_max: 5000 } })
    })
  })

  describe('getById', () => {
    it('should fetch product detail by id', async () => {
      const detail = { id: 1, title: '商品', description: 'desc', price: 1000, status: 1, user_id: 1, category_id: null, created_at: '', updated_at: '', images: [] }
      mockGet.mockResolvedValue({ data: { data: detail } })

      const result = await productAPI.getById(1)
      expect(mockGet).toHaveBeenCalledWith('/products/1')
      expect(result.title).toBe('商品')
    })
  })

  describe('create', () => {
    it('should create a product', async () => {
      const created = { id: 1, title: '新商品', description: 'desc', price: 2000, status: 1, user_id: 1, category_id: null, created_at: '', updated_at: '', images: [] }
      mockPost.mockResolvedValue({ data: { data: created } })

      const result = await productAPI.create({ title: '新商品', description: 'desc', price: 2000 })
      expect(mockPost).toHaveBeenCalledWith('/products', { title: '新商品', description: 'desc', price: 2000 })
      expect(result.id).toBe(1)
    })
  })

  describe('update', () => {
    it('should update a product', async () => {
      const updated = { id: 1, title: '更新', description: '', price: 3000, status: 1, user_id: 1, category_id: null, created_at: '', updated_at: '', images: [] }
      mockPut.mockResolvedValue({ data: { data: updated } })

      const result = await productAPI.update(1, { title: '更新', description: '', price: 3000 })
      expect(mockPut).toHaveBeenCalledWith('/products/1', { title: '更新', description: '', price: 3000 })
      expect(result.title).toBe('更新')
    })
  })

  describe('delete', () => {
    it('should delete a product', async () => {
      mockDelete.mockResolvedValue({ data: { code: 0, message: 'ok' } })

      const result = await productAPI.delete(1)
      expect(mockDelete).toHaveBeenCalledWith('/products/1')
      expect(result.code).toBe(0)
    })
  })

  describe('listMine', () => {
    it('should fetch my products with status filter', async () => {
      mockGet.mockResolvedValue({ data: { data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } })

      await productAPI.listMine({ status: 1, page: 1, size: 20 })
      expect(mockGet).toHaveBeenCalledWith('/products/mine', { params: { status: 1, page: 1, size: 20 } })
    })
  })

  describe('getCounts', () => {
    it('should fetch product counts', async () => {
      mockGet.mockResolvedValue({ data: { data: { active: 10, sold: 3, inactive: 2 } } })

      const result = await productAPI.getCounts()
      expect(mockGet).toHaveBeenCalledWith('/products/mine/counts')
      expect(result.active).toBe(10)
    })
  })

  describe('updateStatus', () => {
    it('should update product status', async () => {
      mockPut.mockResolvedValue({ data: { code: 0, message: 'ok' } })

      const result = await productAPI.updateStatus(1, 0)
      expect(mockPut).toHaveBeenCalledWith('/products/1/status', { status: 0 })
      expect(result.code).toBe(0)
    })
  })

  describe('image operations', () => {
    it('should upload image', async () => {
      const file = new File([''], 'test.jpg', { type: 'image/jpeg' })
      mockPost.mockResolvedValue({ data: { data: { id: 1, url: 'http://minio/img.jpg', sort_order: 0 } } })

      const result = await productAPI.uploadImage(file)
      expect(mockPost).toHaveBeenCalledWith('/images/upload', expect.any(FormData), {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      expect(result.id).toBe(1)
    })

    it('should get image by id', async () => {
      mockGet.mockResolvedValue({ data: { data: { id: 1, url: 'http://minio/img.jpg', sort_order: 0 } } })

      const result = await productAPI.getImage(1)
      expect(mockGet).toHaveBeenCalledWith('/images/1')
      expect(result.url).toContain('minio')
    })

    it('should delete image', async () => {
      mockDelete.mockResolvedValue({ data: { code: 0, message: 'ok' } })

      const result = await productAPI.deleteImage(1)
      expect(mockDelete).toHaveBeenCalledWith('/images/1')
      expect(result.code).toBe(0)
    })
  })
})
