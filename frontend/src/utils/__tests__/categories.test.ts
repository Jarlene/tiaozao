import { describe, it, expect } from 'vitest'
import { flattenCategories } from '../categories'

describe('flattenCategories', () => {
  it('should flatten a flat list of categories', () => {
    const input = [
      { id: 1, name: '电子', parent_id: null },
      { id: 2, name: '服装', parent_id: null },
    ]
    const result = flattenCategories(input)
    expect(result).toEqual([
      { label: '电子', value: 1 },
      { label: '服装', value: 2 },
    ])
  })

  it('should flatten nested categories with prefix', () => {
    const input = [
      {
        id: 1, name: '电子', parent_id: null,
        children: [
          { id: 3, name: '手机', parent_id: 1 },
        ],
      },
      { id: 2, name: '服装', parent_id: null },
    ]
    const result = flattenCategories(input)
    expect(result).toEqual([
      { label: '电子', value: 1 },
      { label: '电子 / 手机', value: 3 },
      { label: '服装', value: 2 },
    ])
  })

  it('should handle deeply nested categories', () => {
    const input = [
      {
        id: 1, name: '电子', parent_id: null,
        children: [
          {
            id: 3, name: '手机', parent_id: 1,
            children: [
              { id: 5, name: 'iPhone', parent_id: 3 },
            ],
          },
        ],
      },
    ]
    const result = flattenCategories(input)
    expect(result).toEqual([
      { label: '电子', value: 1 },
      { label: '电子 / 手机', value: 3 },
      { label: '电子 / 手机 / iPhone', value: 5 },
    ])
  })

  it('should handle empty input', () => {
    expect(flattenCategories([])).toEqual([])
  })

  it('should handle categories without children', () => {
    const input = [{ id: 1, name: '测试', parent_id: null, children: [] }]
    const result = flattenCategories(input)
    expect(result).toEqual([{ label: '测试', value: 1 }])
  })
})
