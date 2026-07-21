interface MainContentProps {
  children: React.ReactNode;
}

export function MainContent({ children }: MainContentProps) {
  return (
    <main className="flex-1 bg-gray-50 dark:bg-gray-950 min-h-screen">
      {children}
    </main>
  );
}
