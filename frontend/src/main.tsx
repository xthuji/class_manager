import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import { Sidebar } from '@/components/Layout/Sidebar';
import { Header } from '@/components/Layout/Header';
import { MainContent } from '@/components/Layout/MainContent';
import { Dashboard } from '@/pages/Dashboard';
import { StudentManagement } from '@/pages/StudentManagement';
import { HourRecordManagement } from '@/pages/HourRecordManagement';
import { CourseManagement } from '@/pages/CourseManagement';
import { ThresholdManagement } from '@/pages/ThresholdManagement';
import { NotificationCenter } from '@/pages/NotificationCenter';
import { BackupManagement } from '@/pages/BackupManagement';
import '@/styles/globals.css';

const pageTitles: Record<string, string> = {
  '/': '仪表盘',
  '/students': '学生管理',
  '/hour-records': '课时记录',
  '/courses': '课程管理',
  '/threshold': '阈值设置',
  '/notifications': '通知中心',
  '/backup': '备份管理',
};

function AppContent() {
  const location = useLocation();
  const currentTitle = pageTitles[location.pathname] || 'Class Manager';

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      <Sidebar />
      <div className="ml-64 transition-all duration-300">
        <Header title={currentTitle} />
        <MainContent>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/students" element={<StudentManagement />} />
            <Route path="/hour-records" element={<HourRecordManagement />} />
            <Route path="/courses" element={<CourseManagement />} />
            <Route path="/threshold" element={<ThresholdManagement />} />
            <Route path="/notifications" element={<NotificationCenter />} />
            <Route path="/backup" element={<BackupManagement />} />
          </Routes>
        </MainContent>
      </div>
    </div>
  );
}

function App() {
  return (
    <Router>
      <AppContent />
    </Router>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);