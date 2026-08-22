<template>
  <UApp>
    <div class="page-shell">
      <UHeader :toggle="false" class="site-header">
        <template #left>
          <a class="brand" href="/" aria-label="Briefly home">
            <span class="brand-mark" aria-hidden="true">B</span>
            <span>Briefly</span>
          </a>
        </template>

        <template #right>
          <nav class="auth-actions" aria-label="Account navigation">
            <UButton
              to="/login"
              color="neutral"
              :variant="$route.path === '/login' ? 'soft' : 'ghost'"
              size="md"
            >
              Log in
            </UButton>
            <UButton to="/signup" color="primary" variant="solid" size="md" class="signup-button">
              Sign up
            </UButton>
          </nav>
        </template>
      </UHeader>

      <main class="auth-main">
        <RouterView />
      </main>
    </div>
  </UApp>
</template>
