import { useState, useEffect } from 'react';
import { Bell, Check, CheckCircle, Clock, Filter } from 'lucide-react';
import { useNotificationStore } from '@/store/notificationStore';

type FilterType = 'all' | 'unread' | 'read' | 'processed';

export function NotificationCenter() {
  const { notifications, loading, getNotifications, markAsRead, batchMarkAsRead, markAsProcessed, batchMarkAsProcessed, getUnreadCount } = useNotificationStore();
  const [filter, setFilter] = useState<FilterType>('all');
  const [selectedIds, setSelectedIds] = useState<number[]>([]);

  useEffect(() => {
    getNotifications('');
    getUnreadCount();
  }, [getNotifications, getUnreadCount]);

  const handleFilterChange = (newFilter: FilterType) => {
    setFilter(newFilter);
    if (newFilter === 'all') {
      getNotifications('');
    } else {
      getNotifications(newFilter);
    }
  };

  const filteredNotifications = notifications.filter((n) => {
    if (filter === 'all') return true;
    return n.status === filter;
  });

  const toggleSelectAll = () => {
    if (selectedIds.length === filteredNotifications.length) {
      setSelectedIds([]);
    } else {
      setSelectedIds(filteredNotifications.map((n) => n.id));
    }
  };

  const handleSelectAll = () => {
    toggleSelectAll();
  };

  const toggleSelect = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter((i) => i !== id));
    } else {
      setSelectedIds([...selectedIds, id]);
    }
  };

  const handleBatchMarkAsRead = () => {
    batchMarkAsRead(selectedIds);
    setSelectedIds([]);
  };

  const handleBatchMarkAsProcessed = () => {
    batchMarkAsProcessed(selectedIds);
    setSelectedIds([]);
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'unread': return '未读';
      case 'read': return '已读';
      case 'processed': return '已处理';
      default: return status;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'unread': return 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400';
      case 'read': return 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400';
      case 'processed': return 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400';
      default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'unread': return <Clock className="w-4 h-4" />;
      case 'read': return <Check className="w-4 h-4" />;
      case 'processed': return <CheckCircle className="w-4 h-4" />;
      default: return <Bell className="w-4 h-4" />;
    }
  };

  const renderNotifications = () => {
    if (loading) {
      return (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      );
    }

    if (filteredNotifications.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
          <Bell className="w-12 h-12 mb-4 opacity-50" />
          <p>暂无通知</p>
        </div>
      );
    }

    return (
      <>
        <div className="p-4 border-b border-gray-200 dark:border-gray-800 flex items-center space-x-3">
          <input
            type="checkbox"
            checked={selectedIds.length === filteredNotifications.length && filteredNotifications.length > 0}
            onChange={handleSelectAll}
            className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700"
          />
          <span className="text-sm text-gray-500 dark:text-gray-400">全选</span>
        </div>
        {filteredNotifications.map((notification) => (
          <div
            key={notification.id}
            className={`p-4 flex items-start space-x-4 transition-colors ${
              selectedIds.includes(notification.id) ? 'bg-primary-50 dark:bg-primary-900/20' : 'hover:bg-gray-50 dark:hover:bg-gray-800'
            }`}
          >
            <input
              type="checkbox"
              checked={selectedIds.includes(notification.id)}
              onChange={() => toggleSelect(notification.id)}
              className="mt-1 w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700"
            />

            <div className={`flex-shrink-0 w-10 h-10 rounded-lg ${getStatusColor(notification.status)} flex items-center justify-center`}>
              {getStatusIcon(notification.status)}
            </div>

            <div className="flex-1">
              <div className="flex items-center space-x-2">
                <h4 className="font-medium text-gray-800 dark:text-white">{notification.student_name}</h4>
                <span className={`text-xs px-2 py-0.5 rounded-full ${getStatusColor(notification.status)}`}>
                  {getStatusLabel(notification.status)}
                </span>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {notification.subject} - 当前剩余 {notification.current_hours.toFixed(1)} 课时，阈值 {notification.threshold} 课时
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                {new Date(notification.created_at).toLocaleString('zh-CN')}
              </p>
            </div>

            <div className="flex items-center space-x-2">
              {notification.status === 'unread' && (
                <button
                  onClick={() => markAsRead(notification.id)}
                  className="px-3 py-1.5 text-sm text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 dark:text-blue-400 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 transition-colors"
                >
                  标记已读
                </button>
              )}
              {notification.status !== 'processed' && (
                <button
                  onClick={() => markAsProcessed(notification.id)}
                  className="px-3 py-1.5 text-sm text-green-600 bg-green-50 rounded-lg hover:bg-green-100 dark:text-green-400 dark:bg-green-900/30 dark:hover:bg-green-900/50 transition-colors"
                >
                  标记已处理
                </button>
              )}
            </div>
          </div>
        ))}
      </>
    );
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">通知中心</h2>
        {selectedIds.length > 0 && (
          <div className="flex items-center space-x-3">
            <span className="text-sm text-gray-600 dark:text-gray-400">已选择 {selectedIds.length} 项</span>
            <button
              onClick={handleBatchMarkAsRead}
              className="flex items-center space-x-1 px-3 py-1.5 text-sm text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 dark:text-blue-400 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 transition-colors"
            >
              <Check className="w-4 h-4" />
              <span>标记已读</span>
            </button>
            <button
              onClick={handleBatchMarkAsProcessed}
              className="flex items-center space-x-1 px-3 py-1.5 text-sm text-green-600 bg-green-50 rounded-lg hover:bg-green-100 dark:text-green-400 dark:bg-green-900/30 dark:hover:bg-green-900/50 transition-colors"
            >
              <CheckCircle className="w-4 h-4" />
              <span>标记已处理</span>
            </button>
          </div>
        )}
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <div className="flex items-center space-x-2">
            <Filter className="w-5 h-5 text-gray-500 dark:text-gray-400" />
            <div className="flex items-center space-x-1">
              {(['all', 'unread', 'read', 'processed'] as FilterType[]).map((f) => (
                <button
                  key={f}
                  onClick={() => handleFilterChange(f)}
                  className={`px-3 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                    filter === f
                      ? 'bg-primary-600 text-white'
                      : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                  }`}
                >
                  {f === 'all' ? '全部' : getStatusLabel(f)}
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {renderNotifications()}
        </div>
      </div>
    </div>
  );
}
