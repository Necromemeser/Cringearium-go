import type { Course } from '../../api/courses'
import CourseCard from './CourseCard'

export default function CourseGrid({ items }: { items: Course[] }) {
  return <div className="course-grid">{items.map((course) => <CourseCard key={course.id} course={course} />)}</div>
}
