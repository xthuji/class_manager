import { useState, useEffect } from 'react';
import { Plus, Search, Edit2, Trash2, Eye, RefreshCw } from 'lucide-react';
import { useStudentStore } from '@/store/studentStore';
import { api } from '@/api/client';
import { Student } from '@/types';
import { Modal } from '@/components/UI/Modal';
import { ConfirmModal } from '@/components/UI/ConfirmModal';
import { DataTable } from '@/components/Data/DataTable';
import { ProgressBar } from '@/components/Data/ProgressBar';

export function StudentManagement() {
  const { students, loading, createStudent, updateStudent, deleteStudent, listStudents, searchStudents } = useStudentStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [studentToDelete, setStudentToDelete] = useState<Student | null>(null);
  const [editingStudent, setEditingStudent] = useState<Student | null>(null);
  const [searchName, setSearchName] = useState('');
  const [searchStudentID, setSearchStudentID] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    student_id: '',
    contact: '',
    subjects: '',
    total_hours: 0,
  });

  useEffect(() => {
    listStudents();
  }, [listStudents]);

  const handleOpenModal = (student?: Student) => {
    if (student) {
      setEditingStudent(student);
      setFormData({
        name: student.name,
        student_id: student.student_id,
        contact: student.contact,
        subjects: student.subjects.join(','),
        total_hours: student.total_hours,
      });
    } else {
      setEditingStudent(null);
      setFormData({
        name: '',
        student_id: '',
        contact: '',
        subjects: '',
        total_hours: 0,
      });
    }
    setIsModalOpen(true);
  };

  const handleSubmit = async () => {
    const subjects = formData.subjects.split(',').map((s) => s.trim()).filter((s) => s);
    
    if (editingStudent) {
      await updateStudent({
        id: editingStudent.id,
        name: formData.name,
        contact: formData.contact,
        subjects,
        total_hours: formData.total_hours,
      });
    } else {
      await createStudent({
        name: formData.name,
        student_id: formData.student_id,
        contact: formData.contact,
        subjects,
        total_hours: formData.total_hours,
      });
    }
    setIsModalOpen(false);
  };

  const handleDelete = async () => {
    if (studentToDelete) {
      await deleteStudent(studentToDelete.id);
      setStudentToDelete(null);
      setIsConfirmOpen(false);
    }
  };

  const handleSearch = () => {
    searchStudents({
      name: searchName,
      student_id: searchStudentID,
      subjects: [],
    });
  };

  const columns: { key: keyof Student; label: string; render?: (value: Student[keyof Student], row: Student) => React.ReactNode }[] = [
    { key: 'name', label: '姓名' },
    { key: 'student_id', label: '学号' },
    { key: 'contact', label: '联系方式' },
    { key: 'subjects', label: '科目', render: (value) => (Array.isArray(value) ? value.join(', ') : String(value)) },
    { key: 'total_hours', label: '总课时' },
    { key: 'completed_hours', label: '已完成' },
    { 
      key: 'remaining_hours', 
      label: '课时进度', 
      render: (_: Student[keyof Student], row: Student) => (
        <ProgressBar 
          value={row.total_hours - row.remaining_hours} 
          max={row.total_hours} 
          color={row.remaining_hours < 5 ? 'danger' : row.remaining_hours < 10 ? 'warning' : 'success'}
          showLabel={false}
        />
      ),
    },
  ];

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">学生管理</h2>
        <div className="flex items-center space-x-3">
          <button
            onClick={() => handleOpenModal()}
            className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            <Plus className="w-5 h-5" />
            <span>添加学生</span>
          </button>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl p-4 mb-6 border border-gray-200 dark:border-gray-800">
        <div className="flex items-center space-x-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="搜索姓名..."
              value={searchName}
              onChange={(e) => setSearchName(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="搜索学号..."
              value={searchStudentID}
              onChange={(e) => setSearchStudentID(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <button
            onClick={handleSearch}
            className="px-4 py-2 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          >
            搜索
          </button>
        </div>
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <DataTable
          data={students}
          columns={columns}
          loading={loading}
          actions={(row) => (
            <>
              <button
                onClick={() => handleOpenModal(row)}
                className="p-2 text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                title="编辑"
              >
                <Edit2 className="w-4 h-4" />
              </button>
              <button
                onClick={() => { setStudentToDelete(row); setIsConfirmOpen(true); }}
                className="p-2 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                title="删除"
              >
                <Trash2 className="w-4 h-4" />
              </button>
              <button
                className="p-2 text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/30 rounded-lg transition-colors"
                title="查看详情"
              >
                <Eye className="w-4 h-4" />
              </button>
            </>
          )}
        />
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingStudent ? '编辑学生' : '添加学生'}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">姓名 *</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">学号</label>
            <div className="flex items-center space-x-2">
              <input
                type="text"
                value={formData.student_id}
                onChange={(e) => setFormData({ ...formData, student_id: e.target.value })}
                className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                placeholder="留空自动生成"
                disabled={!!editingStudent}
              />
              {!editingStudent && (
                <button
                  onClick={async () => {
                    const studentID = await api.student.generateID();
                    setFormData({ ...formData, student_id: studentID as string });
                  }}
                  className="px-3 py-2 text-sm text-gray-600 bg-gray-100 border border-gray-300 rounded-lg hover:bg-gray-200 dark:text-gray-300 dark:bg-gray-800 dark:border-gray-700 dark:hover:bg-gray-700 transition-colors"
                  title="自动生成学号"
                >
                  <RefreshCw className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">联系方式</label>
            <input
              type="text"
              value={formData.contact}
              onChange={(e) => setFormData({ ...formData, contact: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">科目（逗号分隔）</label>
            <input
              type="text"
              value={formData.subjects}
              onChange={(e) => setFormData({ ...formData, subjects: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">初始课时</label>
            <input
              type="number"
              value={formData.total_hours}
              onChange={(e) => setFormData({ ...formData, total_hours: parseFloat(e.target.value) || 0 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              min="0"
              step="0.5"
            />
          </div>
          <div className="flex justify-end space-x-3 pt-4">
            <button
              onClick={() => setIsModalOpen(false)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-900 dark:border-gray-700 dark:hover:bg-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleSubmit}
              className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 transition-colors"
            >
              {editingStudent ? '保存' : '创建'}
            </button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        isOpen={isConfirmOpen}
        onClose={() => setIsConfirmOpen(false)}
        onConfirm={handleDelete}
        title="确认删除"
        message={`确定要删除学生 "${studentToDelete?.name}" 吗？此操作无法撤销。`}
      />
    </div>
  );
}
