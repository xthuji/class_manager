import { useState, useEffect, useMemo } from 'react';
import { Plus, Filter, Calendar, Users, BookOpen } from 'lucide-react';
import { useHourRecordStore } from '@/store/hourRecordStore';
import { useStudentStore } from '@/store/studentStore';
import { useCourseStore } from '@/store/courseStore';
import { HourRecordWithDetails, BatchHourRecordRequest } from '@/types';
import { Modal } from '@/components/UI/Modal';

export function HourRecordManagement() {
  const { records, loading, batchCreateHourRecord, listHourRecords, getDistinctDates } = useHourRecordStore();
  const { students } = useStudentStore();
  const { courses } = useCourseStore();
  const [isBatchModalOpen, setIsBatchModalOpen] = useState(false);
  const [distinctDates, setDistinctDates] = useState<string[]>([]);
  const [selectedStudents, setSelectedStudents] = useState<number[]>([]);
  const [selectedCourses, setSelectedCourses] = useState<number[]>([]);
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [batchFormData, setBatchFormData] = useState<BatchHourRecordRequest>({
    student_ids: [],
    course_id: 0,
    hours: 1,
    record_date: new Date().toISOString().split('T')[0],
    description: '',
  });

  useEffect(() => {
    getDistinctDates().then((dates: string[]) => {
      setDistinctDates(dates);
    });
    loadRecords();
  }, []);

  const loadRecords = () => {
    listHourRecords({
      student_ids: selectedStudents,
      course_ids: selectedCourses,
      start_date: selectedDate,
      end_date: selectedDate,
    });
  };

  const handleFilterChange = () => {
    loadRecords();
  };

  const handleBatchSubmit = async () => {
    await batchCreateHourRecord(batchFormData);
    setIsBatchModalOpen(false);
    loadRecords();
    getDistinctDates().then((dates: string[]) => {
      setDistinctDates(dates);
    });
  };

  const groupedRecords = useMemo(() => {
    const groups: Record<string, Record<string, Record<string, HourRecordWithDetails[]>>> = {};
    
    records.forEach((record) => {
      const date = record.record_date;
      const subject = record.course_subject || '未分类';
      const student = `${record.student_name} (${record.student_no})`;
      
      if (!groups[date]) groups[date] = {};
      if (!groups[date][subject]) groups[date][subject] = {};
      if (!groups[date][subject][student]) groups[date][subject][student] = [];
      groups[date][subject][student].push(record);
    });
    
    return groups;
  }, [records]);

  const allStudents = useMemo(() => {
    const studentMap: Record<string, string> = {};
    records.forEach((record) => {
      const key = `${record.student_name} (${record.student_no})`;
      studentMap[key] = key;
    });
    return Object.keys(studentMap);
  }, [records]);

  

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white">课时记录</h2>
        <button
          onClick={() => setIsBatchModalOpen(true)}
          className="flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus className="w-5 h-5" />
          <span>批量记录课时</span>
        </button>
      </div>

      <div className="bg-white dark:bg-gray-900 rounded-xl p-4 mb-6 border border-gray-200 dark:border-gray-800">
        <div className="flex items-center space-x-2 mb-4">
          <Filter className="w-5 h-5 text-gray-500" />
          <span className="font-medium text-gray-700 dark:text-gray-300">筛选条件</span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="flex items-center space-x-1 text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              <Users className="w-4 h-4" />
              <span>学生</span>
            </label>
            <select
              value={selectedStudents.join(',')}
              onChange={(e) => {
                const ids = e.target.value ? e.target.value.split(',').map(Number) : [];
                setSelectedStudents(ids);
              }}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              multiple
            >
              {students.map((student) => (
                <option key={student.id} value={student.id}>
                  {student.name} ({student.student_id})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="flex items-center space-x-1 text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              <BookOpen className="w-4 h-4" />
              <span>科目</span>
            </label>
            <select
              value={selectedCourses.join(',')}
              onChange={(e) => {
                const ids = e.target.value ? e.target.value.split(',').map(Number) : [];
                setSelectedCourses(ids);
              }}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              multiple
            >
              {courses.map((course) => (
                <option key={course.id} value={course.id}>
                  {course.name} ({course.subject})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="flex items-center space-x-1 text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              <Calendar className="w-4 h-4" />
              <span>日期</span>
            </label>
            <select
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="">全部日期</option>
              {distinctDates.map((date) => (
                <option key={date} value={date}>{date}</option>
              ))}
            </select>
          </div>
        </div>
        <button
          onClick={handleFilterChange}
          className="mt-4 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 transition-colors"
        >
          应用筛选
        </button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      ) : Object.keys(groupedRecords).length === 0 ? (
        <div className="bg-white dark:bg-gray-900 rounded-xl p-8 border border-gray-200 dark:border-gray-800 text-center">
          <p className="text-gray-500 dark:text-gray-400">暂无课时记录</p>
        </div>
      ) : (
        <div className="space-y-6">
          {Object.entries(groupedRecords).map(([date, subjects]) => (
            <div key={date} className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
              <div className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
                <span className="font-semibold text-gray-800 dark:text-white">{date}</span>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="bg-gray-50 dark:bg-gray-800">
                      <th className="px-4 py-2 text-left text-sm font-medium text-gray-600 dark:text-gray-400">学生</th>
                      {Object.keys(subjects).map((subject) => (
                        <th key={subject} className="px-4 py-2 text-center text-sm font-medium text-gray-600 dark:text-gray-400">
                          {subject}
                        </th>
                      ))}
                      <th className="px-4 py-2 text-center text-sm font-medium text-gray-600 dark:text-gray-400">总计</th>
                    </tr>
                  </thead>
                  <tbody>
                    {allStudents.map((student) => {
                      let total = 0;
                      return (
                        <tr key={student} className="border-t border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                          <td className="px-4 py-2 text-sm text-gray-800 dark:text-white">{student}</td>
                          {Object.values(subjects).map((subjectRecords, idx) => {
                            const recordsForStudent = subjectRecords[student] || [];
                            const hours = recordsForStudent.reduce((sum, r) => sum + r.hours, 0);
                            total += hours;
                            return (
                              <td key={idx} className="px-4 py-2 text-center text-sm">
                                {hours > 0 ? (
                                  <span className="inline-flex items-center justify-center w-10 h-10 bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-400 rounded-lg font-medium">
                                    {hours}
                                  </span>
                                ) : (
                                  <span className="text-gray-300">-</span>
                                )}
                              </td>
                            );
                          })}
                          <td className="px-4 py-2 text-center text-sm font-semibold text-gray-800 dark:text-white">
                            {total}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal
        isOpen={isBatchModalOpen}
        onClose={() => setIsBatchModalOpen(false)}
        title="批量记录课时"
        size="lg"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">选择学生（可多选）</label>
            <div className="max-h-40 overflow-y-auto space-y-2">
              {students.map((student) => (
                <label key={student.id} className="flex items-center space-x-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={batchFormData.student_ids.includes(student.id)}
                    onChange={(e) => {
                      const newIds = e.target.checked
                        ? [...batchFormData.student_ids, student.id]
                        : batchFormData.student_ids.filter((id) => id !== student.id);
                      setBatchFormData({ ...batchFormData, student_ids: newIds });
                    }}
                    className="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-gray-600 dark:bg-gray-700"
                  />
                  <span className="text-gray-700 dark:text-gray-300">{student.name} ({student.student_id}) - 剩余课时: {student.remaining_hours}</span>
                </label>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">课程</label>
            <select
              value={batchFormData.course_id}
              onChange={(e) => setBatchFormData({ ...batchFormData, course_id: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value={0}>请选择课程</option>
              {courses.map((course) => (
                <option key={course.id} value={course.id}>{course.name} ({course.subject})</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">日期</label>
            <input
              type="date"
              value={batchFormData.record_date}
              onChange={(e) => setBatchFormData({ ...batchFormData, record_date: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">消耗课时</label>
            <input
              type="number"
              value={batchFormData.hours}
              onChange={(e) => setBatchFormData({ ...batchFormData, hours: parseFloat(e.target.value) || 1 })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              min="0.5"
              step="0.5"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">备注</label>
            <textarea
              value={batchFormData.description}
              onChange={(e) => setBatchFormData({ ...batchFormData, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-800 text-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              rows={3}
            />
          </div>
          <div className="flex justify-end space-x-3 pt-4">
            <button
              onClick={() => setIsBatchModalOpen(false)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 dark:text-gray-300 dark:bg-gray-900 dark:border-gray-700 dark:hover:bg-gray-800 transition-colors"
            >
              取消
            </button>
            <button
              onClick={handleBatchSubmit}
              disabled={batchFormData.student_ids.length === 0 || batchFormData.course_id === 0 || !batchFormData.record_date}
              className="px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              确认记录 ({batchFormData.student_ids.length}人)
            </button>
          </div>
        </div>
      </Modal>
    </div>
  );
}