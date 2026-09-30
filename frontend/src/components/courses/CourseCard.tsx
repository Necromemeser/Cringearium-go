import { formatCoursePrice, type Course } from '../../api/courses'
import Link from '../common/Link'

export default function CourseCard({ course }: { course: Course }) {
  return <article className="course-card">
    <div className="course-cover"><span>{course.theme ? '∑' : '📚'}</span></div>
    <div className="course-body">
      <span className="course-theme">{course.theme || 'Курс'}</span>
      <h3>{course.title}</h3>
      <p>{course.description || 'Описание курса пока не добавлено.'}</p>
      <div className="course-footer"><strong>{formatCoursePrice(course.price)}</strong><Link href={`/courses/${course.id}`} className="small-button">Подробнее</Link></div>
    </div>
  </article>
}
