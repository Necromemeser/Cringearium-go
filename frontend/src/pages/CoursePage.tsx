import { useEffect, useState } from 'react'
import { completePage, enrollCourse, formatCoursePrice, getCourse, getCourseProgress, type CourseDetails } from '../api/courses'
import type { User } from '../api/auth'
import { TOKEN_KEY } from '../constants/auth'
import Link from '../components/common/Link'
import Markdown from '../components/markdown/Markdown'
import TestPage from '../components/TestPage'
import { navigate } from '../utils/navigation'
export default function CoursePage({ courseId, user }