import { ChevronLeft, ChevronRight } from 'lucide-react';

interface Column<T> {
  key: keyof T;
  label: string;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  loading?: boolean;
  selectedIds?: number[];
  onSelect?: (id: number) => void;
  onSelectAll?: () => void;
  hasCheckbox?: boolean;
  pagination?: {
    current: number;
    total: number;
    pageSize: number;
    onPageChange: (page: number) => void;
  };
  actions?: (row: T) => React.ReactNode;
}

export function DataTable<T extends { id: number }>({
  data,
  columns,
  loading,
  selectedIds = [],
  onSelect,
  onSelectAll,
  hasCheckbox = false,
  pagination,
  actions,
}: DataTableProps<T>) {
  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
        <p>暂无数据</p>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm text-left bg-white dark:bg-gray-900 rounded-lg overflow-hidden">
        <thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-800 dark:text-gray-400">
          <tr>
            {hasCheckbox && (
              <th className="px-4 py-3">
                <input
                  type="checkbox"
                  checked={selectedIds.length === data.length && data.length > 0}
                  onChange={onSelectAll}
                  className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700"
                />
              </th>
            )}
            {columns.map((column) => (
              <th key={String(column.key)} className="px-4 py-3">{column.label}</th>
            ))}
            {actions && <th className="px-4 py-3">操作</th>}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
          {data.map((row) => (
            <tr
              key={row.id}
              className={`hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors ${
                selectedIds.includes(row.id) ? 'bg-primary-50 dark:bg-primary-900/20' : ''
              }`}
            >
              {hasCheckbox && (
                <td className="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(row.id)}
                    onChange={() => onSelect?.(row.id)}
                    className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700"
                  />
                </td>
              )}
              {columns.map((column) => (
                <td key={String(column.key)} className="px-4 py-3 text-gray-600 dark:text-gray-300">
                  {column.render ? column.render(row[column.key], row) : String(row[column.key])}
                </td>
              ))}
              {actions && (
                <td className="px-4 py-3">
                  <div className="flex items-center space-x-2">
                    {actions(row)}
                  </div>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>

      {pagination && (
        <div className="flex items-center justify-between mt-4 px-4">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            显示 {(pagination.current - 1) * pagination.pageSize + 1} 到{' '}
            {Math.min(pagination.current * pagination.pageSize, pagination.total)} 条，
            共 {pagination.total} 条
          </p>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => pagination.onPageChange(pagination.current - 1)}
              disabled={pagination.current === 1}
              className="p-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed dark:border-gray-700 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              {pagination.current} / {Math.ceil(pagination.total / pagination.pageSize)}
            </span>
            <button
              onClick={() => pagination.onPageChange(pagination.current + 1)}
              disabled={pagination.current === Math.ceil(pagination.total / pagination.pageSize)}
              className="p-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed dark:border-gray-700 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
