import type { CategoryTreeItem } from '@/api/products'

export function flattenCategories(
  items: CategoryTreeItem[],
  prefix = '',
): Array<{ label: string; value: number }> {
  const result: Array<{ label: string; value: number }> = []
  for (const item of items) {
    const label = prefix ? `${prefix} / ${item.name}` : item.name
    result.push({ label, value: item.id })
    if (item.children && item.children.length > 0) {
      result.push(...flattenCategories(item.children, label))
    }
  }
  return result
}
