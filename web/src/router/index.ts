import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../components/Dashboard.vue'
import SearchSeries from '../components/SearchSeries.vue'
import SeriesDetail from '@/components/SeriesDetail.vue'

const routes = [
  {
    path: '/',
    name: 'library',
    component: Dashboard
  },
  {
    path: '/series/:id',
    name: 'series-detail',
    component: SeriesDetail,
    props: true
  },
  {
    path: '/add',
    name: 'search',
    component: SearchSeries
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
