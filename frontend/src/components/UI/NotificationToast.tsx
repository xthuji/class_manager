import { CheckCircle, XCircle, Info, X } from 'lucide-react';
import { useEffect, useState } from 'react';

interface NotificationToastProps {
  message: string;
  type: 'success' | 'error' | 'info';
  duration?: number;
  onClose: () => void;
}

export function NotificationToast({ message, type, duration = 3000, onClose }: NotificationToastProps) {
  const [visible, setVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setVisible(false);
      setTimeout(onClose, 300);
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  const icons = {
    success: <CheckCircle className="w-5 h-5 text-green-500" />,
    error: <XCircle className="w-5 h-5 text-red-500" />,
    info: <Info className="w-5 h-5 text-blue-500" />,
  };

  const bgColors = {
    success: 'bg-green-50 border-green-200 dark:bg-green-900/30 dark:border-green-800',
    error: 'bg-red-50 border-red-200 dark:bg-red-900/30 dark:border-red-800',
    info: 'bg-blue-50 border-blue-200 dark:bg-blue-900/30 dark:border-blue-800',
  };

  const textColors = {
    success: 'text-green-800 dark:text-green-400',
    error: 'text-red-800 dark:text-red-400',
    info: 'text-blue-800 dark:text-blue-400',
  };

  return (
    <div
      className={`fixed top-4 right-4 z-50 flex items-center space-x-3 px-4 py-3 rounded-lg border shadow-lg transition-all duration-300 ${
        visible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
      } ${bgColors[type]}`}
    >
      {icons[type]}
      <span className={`text-sm font-medium ${textColors[type]}`}>{message}</span>
      <button
        onClick={() => {
          setVisible(false);
          setTimeout(onClose, 300);
        }}
        className="ml-2 p-1 rounded hover:bg-black/10 dark:hover:bg-white/10 transition-colors"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}
