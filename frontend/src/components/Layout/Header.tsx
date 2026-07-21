import { Moon, Sun, Bell, Settings, User } from 'lucide-react';
import { useAppStore } from '@/store/appStore';
import { useNotificationStore } from '@/store/notificationStore';

interface HeaderProps {
  title: string;
}

export function Header({ title }: HeaderProps) {
  const { theme, toggleTheme } = useAppStore();
  const { unreadCount, getUnreadCount } = useNotificationStore();

  return (
    <header className="h-16 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between px-6 sticky top-0 z-40">
      <div className="flex items-center">
        <h2 className="text-lg font-semibold text-gray-800 dark:text-white">{title}</h2>
      </div>

      <div className="flex items-center space-x-4">
        <button
          onClick={toggleTheme}
          className="p-2 rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
          title={theme === 'light' ? '切换到暗色模式' : '切换到亮色模式'}
        >
          {theme === 'light' ? <Moon className="w-5 h-5" /> : <Sun className="w-5 h-5" />}
        </button>

        <button
          onClick={getUnreadCount}
          className="relative p-2 rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
          title="通知"
        >
          <Bell className="w-5 h-5" />
          {unreadCount > 0 && (
            <span className="absolute top-1 right-1 bg-red-500 text-white text-xs font-bold w-4 h-4 rounded-full flex items-center justify-center">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </button>

        <button
          className="p-2 rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
          title="设置"
        >
          <Settings className="w-5 h-5" />
        </button>

        <div className="flex items-center space-x-3 pl-4 border-l border-gray-200 dark:border-gray-800">
          <div className="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
            <User className="w-4 h-4 text-primary-600 dark:text-primary-400" />
          </div>
          <span className="text-sm font-medium text-gray-700 dark:text-gray-300">管理员</span>
        </div>
      </div>
    </header>
  );
}
