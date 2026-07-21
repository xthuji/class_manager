import { useState, useEffect } from 'react';
import { Plus, Users, History, Trash2 } from 'lucide-react';
import { useHourRechargeStore } from '@/store/hourRechargeStore';
import { useStudentStore } from '@/store/studentStore';
import { Modal } from '@/components/UI/Modal';
import { ConfirmModal } from '@/components/UI/ConfirmModal';
import { HourRecharge, HourRechargeCreateRequest } from '@/types';

export function HourRechargeManagement() {
  const { recharges, loading, createRecharge, listRecharges, getRechargesByStudent, deleteRecharge } = useHourRechargeStore();
  const { students } = useStudentStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedStudent, setSelectedStudent] = useState<number>(0);
  const [showRechargeLog, setShowRechargeLog] = useState(false);
  const [courseToDelete, setCourseToDelete] = useState<HourRecharge | null>(null);
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);
  const [formData, setFormData] = useState<HourRechargeCreateRequest>({
    student_id: 0,
    hours: 0,
    recharge_date: new Date().toISOString().split('T')[0],
    description: '',
  });

  useEffect(() => {
    listRecharges();
  }, []);

  const handleSubmit = async () => {
    await createRecharge(formData);
    setIsModalOpen(false);
    listRecharges();
  };

  const handleStudentChange = (studentID: number) => {
    setSelectedStudent(studentID);
    if (studentID > 0) {
      getRechargesByStudent(studentID);
      setShowRechargeLog(true);
    } else {
      setShowRechargeLog(false);
      listRecharges();
    }
  };

  const handleDelete = async () => {
    if (courseToDelete) {
      await deleteRecharge(courseToDelete.id);
      setCourseToDelete(null);
      setIsConfirmOpen(false);
      listRecharges();
    }
  };

  const student = students.find((s) => s.id === formData.student_id);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">课时充值</h2>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          <Plus className="w-5 h-5" />
          <span>充值课时</span>
        </button>
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl p-4 mb-6 border border-gray-200 dark:border-gray-800">
        <div className="flex items-center space-x-4">
          <div className="flex-1">
            <label className="flex items-center space-x-1 text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              <Users className="w-4 h-4" />
              <span>选择学生查看充值记录</span>
            </label>
            <select
              value={selectedStudent}
              onChange={(e) => handleStudentChange(parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value={0}>全部学生</option>
              {students.map((student) => (
                <option key={student.id} value={student.id}>
                  {student.name} ({student.student_id}) - 总课时: {student.total_hours}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {showRechargeLog && selectedStudent > 0 && (
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden mb-6">
          <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center space-x-2">
              <History className="w-5 h-5 text-gray-500" />
              <span className="font-semibold text-gray-800 dark:text-white">充值记录</span>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 dark:bg-gray-800">
                  <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">日期</th>
                  <th className="px-4 py-2 text-right text-sm font-medium text-gray-600 dark:text-gray-400">充值课时</th>
                  <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">备注</th>
                </tr>
              </thead>
              <tbody>
                {recharges.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                      暂无充值记录
                    </td>
                  </tr>
                ) : (
                  recharges.map((recharge) => (
                    <tr key={recharge.id} className="border-t border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                      <td className="px-4 py-2 text-sm text-gray-800 dark:text-white">
                        {recharge.recharge_date || new Date(recharge.created_at).toLocaleDateString('zh-CN')}
                      </td>
                      <td className="px-4 py-2 text-right text-sm font-semibold text-green-600">
                        +{recharge.hours}
                      </td>
                      <td className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">
                        {recharge.description || '-'}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <span className="font-semibold text-gray-800 dark:text-white">充值日志</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 dark:bg-gray-800">
                <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">学生</th>
                <th className="px-4 py-2 text-right text-sm font-medium text-gray-600 dark:text-gray-400">充值课时</th>
                <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">充值日期</th>
                <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">备注</th>
                <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">操作</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto"></div>
                  </td>
                </tr>
              ) : recharges.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                    暂无充值记录
                  </td>
                </tr>
              ) : (
                recharges.map((recharge) => (
                  <tr key={recharge.id} className="border-t border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td className="px-4 py-2 text-sm text-gray-800 dark:text-white">
                      {recharge.student_name ? `${recharge.student_name} (${recharge.student_no})` : '未知学生'}
                    </td>
                    <td className="px-4 py-2 text-right text-sm font-semibold text-green-600">
                      +{recharge.hours}
                    </td>
                    <td className="px-4 py-2 text-sm text-gray-800 dark:text-white">
                      {recharge.recharge_date || new Date(recharge.created_at).toLocaleDateString('zh-CN')}
                    </td>
                    <td className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">
                      {recharge.description || '-'}
                    </td>
                    <td className="px-4 py-2">
                      <button
                        onClick={() => { setCourseToDelete(recharge); setIsConfirmOpen(true); }}
                        className="p-2 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="充值课时"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">选择学生</label>
            <select
              value={formData.student_id}
              onChange={(e) => setFormData({ ...formData, student_id: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value={0}>请选择学生</option>
              {students.map((student) => (
                <option key={student.id} value={student.id}>
                  {student.name} ({student.student_id}) - 当前课时: {student.total_hours}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">充值课时</label>
            <input
              type="number"
              value={formData.hours}
              onChange={(e) => setFormData({ ...formData, hours: parseFloat(e.target.value) || 0 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              min="1"
              step="1"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">充值日期</label>
            <input
              type="date"
              value={formData.recharge_date}
              onChange={(e) => setFormData({ ...formData, recharge_date: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">备注</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
            />
          </div>
          {student && (
            <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                <span className="font-medium">充值后课时: </span>
                <span className="font-semibold text-green-600">{student.total_hours + formData.hours}</span>
              </p>
            </div>
          )}
          <div className="flex justify-end space-x-3 pt-4">
            <button
              onClick={() => setIsModalOpen(false)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-900 dark:border-gray-700 dark:hover:bg-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleSubmit}
              disabled={formData.student_id === 0 || formData.hours <= 0}
              className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
        message="确定要删除这条充值记录吗？此操作无法撤销！"
        confirmText="确认删除"
      />
    </div>
  );
}