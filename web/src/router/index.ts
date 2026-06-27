import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../components/Dashboard.vue'
import SearchSeries from '../components/SearchSeries.vue'
import SeriesDetail from '@/components/SeriesDetail.vue'
import SettingsPage from '@/components/SettingsPage.vue'
import ActivityPage from '@/components/ActivityPage.vue'

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
    },
    {
        path: '/activity',
        name: 'activity',
        component: ActivityPage
    },
    {
        path: '/settings',
        name: 'settings',
        component: SettingsPage
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes,
    scrollBehavior(_to, _from, savedPosition) {
        if (savedPosition) {
            return new Promise((resolve) => {
                setTimeout(() => {
                    resolve(savedPosition)
                }, 300)
            })
        }
        return { top: 0 }
    }
})

export default router
