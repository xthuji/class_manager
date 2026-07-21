interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  color?: 'blue' | 'green' | 'purple' | 'orange';
}

export function StatCard({ title, value, icon, color = 'blue' }: StatCardProps) {
  const colorClasses = {
    blue: {
      bg: 'bg-blue-50',
      iconBg: 'bg-blue-100 dark:bg-blue-900/30',
      iconColor: 'text-blue-600 dark:text-blue-400',
      textColor: 'text-blue-600',
    },
    green: {
      bg: 'bg-green-50',
      iconBg: 'bg-green-100 dark:bg-green-900/30',
      iconColor: 'text-green-600 dark:text-green-400',
      textColor: 'text-green-600',
    },
    purple: {
      bg: 'bg-purple-50',
      iconBg: 'bg-purple-100 dark:bg-purple-900/30',
      iconColor: 'text-purple-600 dark:text-purple-400',
      textColor: 'text-purple-600',
    },
    orange: {
      bg: 'bg-orange-50',
      iconBg: 'bg-orange-100 dark:bg-orange-900/30',
      iconColor: 'text-orange-600 dark:text-orange-400',
      textColor: 'text-orange-600',
    },
  };

  const classes = colorClasses[color];

  return (
    <div className={`${classes.bg} dark:bg-gray-900 rounded-xl p-5 border border-gray-100 dark:border-gray-800`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500 dark:text-gray-400">{title}</p>
          <p className={`text-2xl font-bold text-gray-800 dark:text-white mt-1`}>{value}</p>
        </div>
        <div className={`${classes.iconBg} p-3 rounded-lg`}>
          <div className={classes.iconColor}>{icon}</div>
        </div>
      </div>
    </div>
  );
}
