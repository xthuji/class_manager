import { Users, BookOpen, Clock, Bell } from 'lucide-react';
import { useEffect } from 'react';
import { useAppStore } from '@/store/appStore';
import { useNotificationStore } from '@/store/notificationStore';
import { StatCard } from '@/components/Data/StatCard';

export function Dashboard() {
  const { stats, loadStats } = useAppStore();
  const { getUnreadCount } = useNotificationStore();

  useEffect(() => {
    loadStats();
    getUnreadCount();
  }, [loadStats, getUnreadCount]);

  return (
    <div className="p-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          title="学生总数"
          value={stats?.total_students || 0}
          icon={<Users className="w-6 h-6" />}
          color="blue"
        />
        <StatCard
          title="课程总数"
          value={stats?.total_courses || 0}
          icon={<BookOpen className="w-6 h-6" />}
          color="green"
        />
        <StatCard
          title="今日课时"
          value={`${stats?.today_hours || 0} 小时`}
          icon={<Clock className="w-6 h-6" />}
          color="purple"
        />
        <StatCard
          title="待处理提醒"
          value={stats?.unread_notifications || 0}
          icon={<Bell className="w-6 h-6" />}
          color="orange"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">最近通知</h3>
          <div className="space-y-3">
            <div className="flex items-center space-x-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <div className="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
                <Bell className="w-5 h-5 text-amber-600 dark:text-amber-400" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-800 dark:text-white">暂无通知</p>
                <p className="text-xs text-gray-500 dark:text-gray-400">最近没有新的提醒</p>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-xl p-6 border border-gray-200 dark:border-gray-800">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">系统状态</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600 dark:text-gray-400">数据库连接</span>
              <span className="flex items-center space-x-2">
                <span className="w-2 h-2 rounded-full bg-green-500"></span>
                <span className="text-sm font-medium text-green-600 dark:text-green-400">正常</span>
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600 dark:text-gray-400">自动备份</span>
              <span className="flex items-center space-x-2">
                <span className="w-2 h-2 rounded-full bg-green-500"></span>
                <span className="text-sm font-medium text-green-600 dark:text-green-400">已启用</span>
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-600 dark:text-gray-400">阈值检查</span>
              <span className="flex items-center space-x-2">
                <span className="w-2 h-2 rounded-full bg-green-500"></span>
                <span className="text-sm font-medium text-green-600 dark:text-green-400">每小时</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
