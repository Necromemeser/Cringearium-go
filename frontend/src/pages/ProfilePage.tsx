import { useEffect, useState } from 'react'
import { getEnrolledCourses, type Course } from '../api/courses'
import type { User } from '../api/auth'
import { TOKEN_KEY } from '../constants/auth'
import Link from '../components/common/Link'
import CourseGrid from '../components/courses/CourseGrid'
export default function ProfilePage({ user }