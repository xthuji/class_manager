interface ProgressBarProps {
  value: number;
  max: number;
  color?: 'primary' | 'success' | 'warning' | 'danger';
  showLabel?: boolean;
}

export function ProgressBar({ value, max, color = 'primary', showLabel = true }: ProgressBarProps) {
  const percentage = max > 0 ? Math.min((value / max) * 100, 100) : 0;

  const colorClasses = {
    primary: 'bg-primary-600',
    success: 'bg-green-600',
    warning: 'bg-amber-500',
    danger: 'bg-red-600',
  };

  const bgColorClasses = {
    primary: 'bg-primary-100 dark:bg-primary-900/30',
    success: 'bg-green-100 dark:bg-green-900/30',
    warning: 'bg-amber-100 dark:bg-amber-900/30',
    danger: 'bg-red-100 dark:bg-red-900/30',
  };

  return (
    <div className="flex items-center space-x-3">
      <div className={`flex-1 h-2 rounded-full ${bgColorClasses[color]} overflow-hidden`}>
        <div
          className={`h-full ${colorClasses[color]} rounded-full transition-all duration-500`}
          style={{ width: `${percentage}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">
          {value.toFixed(1)} / {max.toFixed(1)}
        </span>
      )}
    </div>
  );
}
