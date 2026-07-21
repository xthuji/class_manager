import { useState, useEffect } from 'react';
import { Settings, Save, Download, Upload, Eye } from 'lucide-react';
import { useStudentStore } from '@/store/studentStore';
import { Modal } from '@/components/UI/Modal';

interface ThresholdItem {
  subject: string;
  value: number;
}

export function ThresholdManagement() {
  const { students } = useStudentStore();
  const [thresholds, setThresholds] = useState<ThresholdItem[]>([
    { subject: '数学', value: 5 },
    { subject: '英语', value: 5 },
    { subject: '物理', value: 5 },
    { subject: '化学', value: 5 },
  ]);
  const [previewSubject, setPreviewSubject] = useState('');
  const [previewValue, setPreviewValue] = useState(0);
  const [previewData, setPreviewData] = useState<{ triggerCount: number; totalCount: number } | null>(null);
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);
  const [templateName, setTemplateName] = useState('');
  const [isSaveTemplateModalOpen, setIsSaveTemplateModalOpen] = useState(false);

  useEffect(() => {
    extractSubjectsFromStudents();
  }, [students]);

  const extractSubjectsFromStudents = () => {
    const subjectSet = new Set<string>();
    students.forEach((student) => {
      if (Array.isArray(student.subjects)) {
        student.subjects.forEach((subject) => subjectSet.add(subject));
      }
    });

    const currentSubjects = new Set(thresholds.map((t) => t.subject));
    subjectSet.forEach((subject) => {
      if (!currentSubjects.has(subject)) {
        setThresholds((prev) => [...prev, { subject, value: 5 }]);
      }
    });
  };

  const handleThresholdChange = (subject: string, value: number) => {
    setThresholds((prev) =>
      prev.map((t) => (t.subject === subject ? { ...t, value } : t))
    );
  };

  const handlePreview = (subject: string, value: number) => {
    setPreviewSubject(subject);
    setPreviewValue(value);
    
    const subjectStudents = students.filter((s) => 
      Array.isArray(s.subjects) && s.subjects.includes(subject)
    );
    
    const triggerStudents = subjectStudents.filter((s) => 
      s.remaining_hours < value
    );
    
    setPreviewData({
      triggerCount: triggerStudents.length,
      totalCount: subjectStudents.length,
    });
    setIsPreviewModalOpen(true);
  };

  const handleSaveTemplate = () => {
    localStorage.setItem(`threshold_template_${templateName}`, JSON.stringify(thresholds));
    setIsSaveTemplateModalOpen(false);
    setTemplateName('');
  };

  const handleLoadTemplate = (name: string) => {
    const saved = localStorage.getItem(`threshold_template_${name}`);
    if (saved) {
      setThresholds(JSON.parse(saved));
    }
  };

  const getTemplateNames = () => {
    const names: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith('threshold_template_')) {
        names.push(key.replace('threshold_template_', ''));
      }
    }
    return names;
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">阈值设置</h2>
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setIsSaveTemplateModalOpen(true)}
            className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
          >
            <Save className="w-5 h-5" />
            <span>保存模板</span>
          </button>
          <button
            className="flex items-center space-x-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Download className="w-5 h-5" />
            <span>应用模板</span>
          </button>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <h3 className="font-semibold text-gray-800 dark:text-white">科目阈值设置</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            当学生某科目的剩余课时低于设定阈值时，系统会自动发送提醒通知
          </p>
        </div>

        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {thresholds.map((threshold) => (
            <div key={threshold.subject} className="p-4 flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div className="w-10 h-10 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center">
                  <Settings className="w-5 h-5 text-primary-600 dark:text-primary-400" />
                </div>
                <div>
                  <p className="font-medium text-gray-800 dark:text-white">{threshold.subject}</p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">剩余课时提醒阈值</p>
                </div>
              </div>

              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <input
                    type="number"
                    value={threshold.value}
                    onChange={(e) => handleThresholdChange(threshold.subject, parseFloat(e.target.value) || 0)}
                    className="w-20 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 text-center"
                    min="0"
                  />
                  <span className="text-gray-600 dark:text-gray-400">课时</span>
                </div>

                <button
                  onClick={() => handlePreview(threshold.subject, threshold.value)}
                  className="flex items-center space-x-1 px-3 py-1.5 text-sm text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 dark:text-blue-400 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 transition-colors"
                >
                  <Eye className="w-4 h-4" />
                  <span>预览</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="mt-6 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div className="p-4 border-b border-gray-200 dark:border-gray-800">
          <h3 className="font-semibold text-gray-800 dark:text-white">模板管理</h3>
        </div>
        <div className="p-4">
          {getTemplateNames().length > 0 ? (
            <div className="space-y-2">
              {getTemplateNames().map((name) => (
                <button
                  key={name}
                  onClick={() => handleLoadTemplate(name)}
                  className="w-full flex items-center justify-between p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                >
                  <span className="text-gray-800 dark:text-white">{name}</span>
                  <Upload className="w-4 h-4 text-gray-500 dark:text-gray-400" />
                </button>
              ))}
            </div>
          ) : (
            <p className="text-gray-500 dark:text-gray-400 text-center py-4">暂无保存的模板</p>
          )}
        </div>
      </div>

      <Modal
        isOpen={isPreviewModalOpen}
        onClose={() => setIsPreviewModalOpen(false)}
        title="阈值预览"
      >
        <div className="space-y-4">
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
            <p className="text-sm text-gray-600 dark:text-gray-400">科目</p>
            <p className="text-lg font-semibold text-gray-800 dark:text-white">{previewSubject}</p>
          </div>
          <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
            <p className="text-sm text-gray-600 dark:text-gray-400">设置阈值</p>
            <p className="text-lg font-semibold text-gray-800 dark:text-white">{previewValue} 课时</p>
          </div>
          <div className="p-4 bg-amber-50 dark:bg-amber-900/30 rounded-lg">
            <p className="text-sm text-amber-600 dark:text-amber-400">预计触发提醒</p>
            <p className="text-lg font-semibold text-amber-800 dark:text-amber-300">
              {previewData?.triggerCount || 0} / {previewData?.totalCount || 0} 名学生
            </p>
          </div>
          <button
            onClick={() => setIsPreviewModalOpen(false)}
            className="w-full px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 transition-colors"
          >
            关闭
          </button>
        </div>
      </Modal>

      <Modal
        isOpen={isSaveTemplateModalOpen}
        onClose={() => setIsSaveTemplateModalOpen(false)}
        title="保存阈值模板"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">模板名称</label>
            <input
              type="text"
              value={templateName}
              onChange={(e) => setTemplateName(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              placeholder="例如：标准阈值模板"
            />
          </div>
          <div className="flex justify-end space-x-3">
            <button
              onClick={() => setIsSaveTemplateModalOpen(false)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-900 dark:border-gray-700 dark:hover:bg-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleSaveTemplate}
              disabled={!templateName.trim()}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              保存
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
