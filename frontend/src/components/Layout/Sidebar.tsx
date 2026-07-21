import { 
  LayoutDashboard, 
  Users, 
  Clock, 
  BookOpen, 
  AlertTriangle, 
  Bell, 
  Database,
  ChevronLeft,
  ChevronRight
} from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useNotificationStore } from '@/store/notificationStore';

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { unreadCount } = useNotificationStore();

  const menuItems = [
    { path: '/', label: '仪表盘', icon: LayoutDashboard },
    { path: '/students', label: '学生管理', icon: Users },
    { path: '/hour-records', label: '课时记录', icon: Clock },
    { path: '/courses', label: '课程管理', icon: BookOpen },
    { path: '/threshold', label: '阈值设置', icon: AlertTriangle },
    { path: '/notifications', label: '通知中心', icon: Bell },
    { path: '/backup', label: '备份管理', icon: Database },
  ];

  const handleNavigate = (path: string) => {
    navigate(path);
  };

  return (
    <aside
      className={`fixed left-0 top-0 h-full bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transition-all duration-300 z-50 ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      <div className="flex flex-col h-full">
        <div className={`flex items-center h-16 px-4 border-b border-gray-200 dark:border-gray-800 ${collapsed ? 'justify-center' : ''}`}>
          {!collapsed && (
            <h1 className="text-lg font-bold text-gray-800 dark:text-white">Class Manager</h1>
          )}
          {collapsed && (
            <span className="text-xl font-bold text-primary-600">CM</span>
          )}
        </div>

        <nav className="flex-1 py-4">
          <ul className="space-y-1 px-2">
            {menuItems.map((item) => {
              const Icon = item.icon;
              const isActive = location.pathname === item.path;
              const isNotification = item.path === '/notifications';

              return (
                <li key={item.path}>
                  <button
                    onClick={() => handleNavigate(item.path)}
                    className={`w-full flex items-center px-3 py-2.5 rounded-lg transition-colors duration-200 relative ${
                      isActive
                        ? 'bg-primary-50 text-primary-700 dark:bg-primary-600/20 dark:text-primary-400'
                        : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                    }`}
                  >
                    <Icon className={`w-5 h-5 flex-shrink-0 ${isActive ? 'text-primary-600' : ''}`} />
                    {!collapsed && (
                      <span className="ml-3 text-sm font-medium">{item.label}</span>
                    )}
                    {isNotification && unreadCount > 0 && (
                      <span className="absolute right-3 bg-red-500 text-white text-xs font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center">
                        {unreadCount > 9 ? '9+' : unreadCount}
                      </span>
                    )}
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="p-2 border-t border-gray-200 dark:border-gray-800">
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="w-full flex items-center justify-center px-3 py-2.5 rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 transition-colors"
          >
            {collapsed ? <ChevronRight className="w-5 h-5" /> : <ChevronLeft className="w-5 h-5" />}
          </button>
        </div>
      </div>
    </aside>
  );
}