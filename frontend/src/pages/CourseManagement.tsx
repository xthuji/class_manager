import { useState, useEffect } from 'react';
import { Plus, Edit2, Trash2, Users, Wallet } from 'lucide-react';
import { useCourseStore } from '@/store/courseStore';
import { useStudentStore } from '@/store/studentStore';
import { Course, Student } from '@/types';
import { Modal } from '@/components/UI/Modal';
import { ConfirmModal } from '@/components/UI/ConfirmModal';
import { DataTable } from '@/components/Data/DataTable';

export function CourseManagement() {
  const { courses, loading, createCourse, updateCourse, deleteCourse, listCourses, enrollStudent, getCourseStudents, courseStudents, addCourseHours } = useCourseStore();
  const { students } = useStudentStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [courseToDelete, setCourseToDelete] = useState<Course | null>(null);
  const [editingCourse, setEditingCourse] = useState<Course | null>(null);
  const [selectedCourse, setSelectedCourse] = useState<Course | null>(null);
  const [isEnrollModalOpen, setIsEnrollModalOpen] = useState(false);
  const [isAddHoursModalOpen, setIsAddHoursModalOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    subject: '',
    description: '',
  });
  const [hoursToAdd, setHoursToAdd] = useState(0);

  useEffect(() => {
    listCourses();
  }, [listCourses]);

  const handleOpenModal = (course?: Course) => {
    if (course) {
      setEditingCourse(course);
      setFormData({
        name: course.name,
        subject: course.subject,
        description: course.description,
      });
    } else {
      setEditingCourse(null);
      setFormData({ name: '', subject: '', description: '' });
    }
    setIsModalOpen(true);
  };

  const handleSubmit = async () => {
    if (editingCourse) {
      await updateCourse({
        id: editingCourse.id,
        ...formData,
      });
    } else {
      await createCourse(formData);
    }
    setIsModalOpen(false);
  };

  const handleDelete = async () => {
    if (courseToDelete) {
      await deleteCourse(courseToDelete.id);
      setCourseToDelete(null);
      setIsConfirmOpen(false);
    }
  };

  const handleEnroll = async (studentId: number) => {
    if (selectedCourse) {
      await enrollStudent({ student_id: studentId, course_id: selectedCourse.id });
      await getCourseStudents(selectedCourse.id);
    }
  };

  const handleAddHours = async () => {
    if (selectedCourse) {
      await addCourseHours({ course_id: selectedCourse.id, hours: hoursToAdd });
      setIsAddHoursModalOpen(false);
      await listCourses();
    }
  };

  const columns: { key: keyof Course; label: string }[] = [
    { key: 'name', label: '课程名称' },
    { key: 'subject', label: '科目' },
    { key: 'student_count', label: '学生数' },
    { key: 'description', label: '描述' },
  ];

  const studentColumns: { key: keyof Student; label: string }[] = [
    { key: 'name', label: '姓名' },
    { key: 'student_id', label: '学号' },
    { key: 'contact', label: '联系方式' },
    { key: 'remaining_hours', label: '剩余课时' },
  ];

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">课程管理</h2>
        <button
          onClick={() => handleOpenModal()}
          className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus className="w-5 h-5" />
          <span>添加课程</span>
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-800">
            <h3 className="font-semibold text-gray-800 dark:text-white">课程列表</h3>
          </div>
          <DataTable
            data={courses}
            columns={columns}
            loading={loading}
            actions={(row) => (
              <>
                <button
                  onClick={() => handleOpenModal(row)}
                  className="p-2 text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => { setCourseToDelete(row); setIsConfirmOpen(true); }}
                  className="p-2 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => { setSelectedCourse(row); getCourseStudents(row.id); }}
                  className="p-2 text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/30 rounded-lg transition-colors"
                >
                  <Users className="w-4 h-4" />
                </button>
              </>
            )}
          />
        </div>

        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
            <h3 className="font-semibold text-gray-800 dark:text-white">
              {selectedCourse ? `${selectedCourse.name} - 学生列表` : '选择课程查看学生'}
            </h3>
            {selectedCourse && (
              <div className="flex items-center space-x-2">
                <button
                  onClick={() => setIsAddHoursModalOpen(true)}
                  className="flex items-center space-x-1 px-3 py-1.5 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                >
                  <Wallet className="w-4 h-4" />
                  <span>充值课时</span>
                </button>
                <button
                  onClick={() => setIsEnrollModalOpen(true)}
                  className="flex items-center space-x-1 px-3 py-1.5 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
                >
                  <Plus className="w-4 h-4" />
                  <span>添加学生</span>
                </button>
              </div>
            )}
          </div>
          <div className="p-4">
            {selectedCourse ? (
              <DataTable
                data={courseStudents}
                columns={studentColumns}
                loading={loading}
                actions={() => (
                  <span className="text-gray-500 text-sm">已选课</span>
                )}
              />
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
                <Users className="w-12 h-12 mb-4 opacity-50" />
                <p>请从左侧选择一门课程</p>
              </div>
            )}
          </div>
        </div>
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingCourse ? '编辑课程' : '添加课程'}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">课程名称 *</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">科目</label>
            <input
              type="text"
              value={formData.subject}
              onChange={(e) => setFormData({ ...formData, subject: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">描述</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
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
              {editingCourse ? '保存' : '创建'}
            </button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={isEnrollModalOpen}
        onClose={() => setIsEnrollModalOpen(false)}
        title="添加学生到课程"
      >
        <div className="space-y-3">
          {students.map((student) => {
            const isEnrolled = courseStudents.some((cs) => cs.id === student.id);
            return (
              <button
                key={student.id}
                onClick={() => !isEnrolled && handleEnroll(student.id)}
                disabled={isEnrolled}
                className={`w-full flex items-center justify-between p-3 rounded-lg border transition-colors ${
                  isEnrolled
                    ? 'bg-gray-100 border-gray-200 dark:bg-gray-800 dark:border-gray-700 cursor-not-allowed'
                    : 'bg-white border-gray-200 dark:bg-gray-900 dark:border-gray-700 hover:bg-primary-50 dark:hover:bg-primary-900/20'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <span className="text-gray-800 dark:text-white">{student.name}</span>
                  <span className="text-sm text-gray-500 dark:text-gray-400">{student.student_id}</span>
                </div>
                <span className={`text-sm ${isEnrolled ? 'text-gray-400' : 'text-primary-600 dark:text-primary-400'}`}>
                  {isEnrolled ? '已加入' : '添加'}
                </span>
              </button>
            );
          })}
        </div>
      </Modal>

      <Modal
        isOpen={isAddHoursModalOpen}
        onClose={() => setIsAddHoursModalOpen(false)}
        title={`为 ${selectedCourse?.name} 充值课时`}
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">充值课时数</label>
            <input
              type="number"
              value={hoursToAdd}
              onChange={(e) => setHoursToAdd(parseFloat(e.target.value) || 0)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              min="0.5"
              step="0.5"
            />
          </div>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            此操作将为课程中的所有学生增加 {hoursToAdd} 课时
          </p>
          <div className="flex justify-end space-x-3 pt-4">
            <button
              onClick={() => setIsAddHoursModalOpen(false)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-900 dark:border-gray-700 dark:hover:bg-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleAddHours}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 transition-colors"
            >
              确认充值
            </button>
          </div>
        </div>
      </Modal>

      <ConfirmModal
        isOpen={isConfirmOpen}
        onClose={() => setIsConfirmOpen(false)}
        onConfirm={handleDelete}
        title="确认删除"
        message={`确定要删除课程 "${courseToDelete?.name}" 吗？此操作将同时删除相关的学生选课记录和课时安排。`}
      />
    </div>
  );
}
