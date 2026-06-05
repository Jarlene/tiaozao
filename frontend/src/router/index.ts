import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'products',
        name: 'ProductList',
        component: () => import('@/views/product/ProductList.vue'),
      },
      {
        path: 'products/create',
        name: 'ProductCreate',
        component: () => import('@/views/product/ProductCreate.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'products/mine',
        name: 'MyProducts',
        component: () => import('@/views/product/MyProducts.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'products/:id',
        name: 'ProductDetail',
        component: () => import('@/views/product/ProductDetail.vue'),
      },
      {
        path: 'products/:id/edit',
        name: 'ProductEdit',
        component: () => import('@/views/product/ProductEdit.vue'),
        meta: { requiresAuth: true },
      },
      {
        path: 'categories',
        name: 'CategoryManagement',
        component: () => import('@/views/category/CategoryManagement.vue'),
        meta: { requiresAuth: true },
      },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' })
  } else if ((to.name === 'Login' || to.name === 'Register') && authStore.isAuthenticated) {
    next({ name: 'Home' })
  } else {
    next()
  }
})

export default router
