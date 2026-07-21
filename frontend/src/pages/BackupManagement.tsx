import { useState } from 'react';
import { Download, Upload, AlertTriangle, FileJson, RefreshCw } from 'lucide-react';
import { Modal } from '@/components/UI/Modal';
import { ConfirmModal } from '@/components/UI/ConfirmModal';

export function BackupManagement() {
  const [exportStatus, setExportStatus] = useState<'idle' | 'exporting' | 'done'>('idle');
  const [importStatus, setImportStatus] = useState<'idle' | 'importing' | 'done'>('idle');
  const [importFile, setImportFile] = useState<File | null>(null);
  const [isImportConfirmOpen, setIsImportConfirmOpen] = useState(false);
  const [isSuccessModalOpen, setIsSuccessModalOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  const handleExport = async () => {
    setExportStatus('exporting');
    try {
      const jsonData = await window.go.ExportAllData() as string;
      const blob = new Blob([jsonData], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `class_manager_backup_${new Date().toISOString().split('T')[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      setExportStatus('done');
      setTimeout(() => setExportStatus('idle'), 3000);
    } catch (error) {
      console.error('Export failed:', error);
      setExportStatus('idle');
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setImportFile(file);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;
    
    setIsImportConfirmOpen(true);
  };

  const handleConfirmImport = async () => {
    if (!importFile) return;
    
    setImportStatus('importing');
    try {
      const text = await importFile.text();
      await window.go.ImportAllData(text);
      setImportStatus('done');
      setSuccessMessage('数据导入成功！');
      setIsSuccessModalOpen(true);
      setImportFile(null);
      setTimeout(() => setImportStatus('idle'), 3000);
    } catch (error) {
      console.error('Import failed:', error);
      setImportStatus('idle');
    } finally {
      setIsImportConfirmOpen(false);
    }
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">数据管理</h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-800">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                <Download className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-800 dark:text-white">数据导出</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">导出所有数据为 JSON 文件</p>
              </div>
            </div>
          </div>
          <div className="p-6">
            <div className="flex items-center space-x-4 mb-4">
              <FileJson className="w-8 h-8 text-gray-400" />
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">文件格式</p>
                <p className="font-medium text-gray-800 dark:text-white">JSON</p>
              </div>
            </div>
            <div className="flex items-center space-x-4 mb-6">
              <Download className="w-8 h-8 text-gray-400" />
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">包含数据</p>
                <p className="font-medium text-gray-800 dark:text-white">学生、课程、课时记录、充值记录等</p>
              </div>
            </div>
            <button
              onClick={handleExport}
              disabled={exportStatus === 'exporting'}
              className="w-full flex items-center justify-center space-x-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {exportStatus === 'exporting' ? (
                <>
                  <RefreshCw className="w-5 h-5 animate-spin" />
                  <span>导出中...</span>
                </>
              ) : exportStatus === 'done' ? (
                <>
                  <Download className="w-5 h-5" />
                  <span>导出成功</span>
                </>
              ) : (
                <>
                  <Download className="w-5 h-5" />
                  <span>导出数据</span>
                </>
              )}
            </button>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-800">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
                <Upload className="w-5 h-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <h3 className="font-semibold text-gray-800 dark:text-white">数据导入</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">导入 JSON 文件恢复数据</p>
              </div>
            </div>
          </div>
          <div className="p-6">
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">选择文件</label>
              <div className="flex items-center justify-center w-full">
                <label className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                  <div className="flex flex-col items-center justify-center pt-5 pb-6">
                    <Upload className="w-8 h-8 text-gray-400 mb-2" />
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {importFile ? importFile.name : '点击或拖拽文件到此处'}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">JSON 文件</p>
                  </div>
                  <input
                    type="file"
                    accept=".json"
                    onChange={handleFileChange}
                    className="hidden"
                  />
                </label>
              </div>
            </div>
            <div className="p-4 bg-amber-50 dark:bg-amber-900/30 rounded-lg mb-4">
              <div className="flex items-start space-x-3">
                <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-amber-800 dark:text-amber-300">警告</p>
                  <p className="text-sm text-amber-600 dark:text-amber-400 mt-1">
                    导入数据将覆盖当前数据库中的所有数据。此操作无法撤销，请确保已备份当前数据。
                  </p>
                </div>
              </div>
            </div>
            <button
              onClick={handleImport}
              disabled={!importFile || importStatus === 'importing'}
              className="w-full flex items-center justify-center space-x-2 px-4 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {importStatus === 'importing' ? (
                <>
                  <RefreshCw className="w-5 h-5 animate-spin" />
                  <span>导入中...</span>
                </>
              ) : importStatus === 'done' ? (
                <>
                  <Upload className="w-5 h-5" />
                  <span>导入成功</span>
                </>
              ) : (
                <>
                  <Upload className="w-5 h-5" />
                  <span>导入数据</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      <ConfirmModal
        isOpen={isImportConfirmOpen}
        onClose={() => setIsImportConfirmOpen(false)}
        onConfirm={handleConfirmImport}
        title="确认导入"
        message={`确定要导入文件 "${importFile?.name}" 吗？此操作将覆盖当前所有数据，且无法撤销！`}
        confirmText="确认导入"
      />

      <Modal
        isOpen={isSuccessModalOpen}
        onClose={() => setIsSuccessModalOpen(false)}
        title="操作成功"
      >
        <div className="text-center py-6">
          <div className="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <p className="text-lg font-medium text-gray-800 dark:text-white">{successMessage}</p>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">页面将自动刷新以显示最新数据</p>
          <button
            onClick={() => {
              setIsSuccessModalOpen(false);
              window.location.reload();
            }}
            className="mt-4 px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            确定
          </button>
        </div>
      </Modal>
    </div>
  );
}