// API Client
const API_BASE_URL = '/api';
let authToken = localStorage.getItem('token');
let currentUser = null;

const api = {
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    if (authToken) {
      headers['Authorization'] = `Bearer ${authToken}`;
    }
    return headers;
  },

  async request(endpoint, options = {}) {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers,
      },
    });

    if (response.status === 401) {
      this.logout();
      window.location.href = '/login.html';
      return null;
    }

    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || 'API error');
    }
    return data;
  },

  // Auth
  async signup(email, password, fullName, timezone = 'UTC') {
    return this.request('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ email, password, full_name: fullName, timezone }),
    });
  },

  async login(email, password) {
    return this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  async getCurrentUser() {
    return this.request('/auth/me', { method: 'GET' });
  },

  async requestPasswordReset(email) {
    return this.request('/auth/request-reset', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
  },

  async resetPassword(token, newPassword) {
    return this.request('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, new_password: newPassword }),
    });
  },

  // Meals
  async createMeal(mealType, date, time, portion, notes = '') {
    return this.request('/meals', {
      method: 'POST',
      body: JSON.stringify({ meal_type: mealType, date, time, portion, notes }),
    });
  },

  async getTodaysMeals() {
    return this.request('/meals/today', { method: 'GET' });
  },

  async getMealsByDateRange(startDate, endDate) {
    return this.request(`/meals/range?start_date=${startDate}&end_date=${endDate}`, { method: 'GET' });
  },

  async updateMeal(id, time, portion, notes) {
    return this.request(`/meals/update?id=${id}`, {
      method: 'PUT',
      body: JSON.stringify({ time, portion, notes }),
    });
  },

  async deleteMeal(id) {
    return this.request(`/meals/delete?id=${id}`, { method: 'DELETE' });
  },

  // Admin
  async getSettings() {
    return this.request('/admin/settings', { method: 'GET' });
  },

  async updateSettings(settings) {
    return this.request('/admin/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    });
  },

  async listUsers() {
    return this.request('/admin/users', { method: 'GET' });
  },

  async updateUserRole(id, isAdmin) {
    return this.request(`/admin/users/role?id=${id}`, {
      method: 'PUT',
      body: JSON.stringify({ is_admin: isAdmin }),
    });
  },

  async updateUserTimezone(id, timezone) {
    return this.request(`/admin/users/timezone?id=${id}`, {
      method: 'PUT',
      body: JSON.stringify({ timezone }),
    });
  },

  async deleteUser(id) {
    return this.request(`/admin/users?id=${id}`, { method: 'DELETE' });
  },

  async resetUserPassword(id, newPassword) {
    return this.request(`/admin/users/reset-password?id=${id}`, {
      method: 'POST',
      body: JSON.stringify({ new_password: newPassword }),
    });
  },

  setToken(token) {
    authToken = token;
    localStorage.setItem('token', token);
  },

  logout() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  isAuthenticated() {
    return !!authToken;
  },

  async ensureAuthenticated() {
    if (!authToken) {
      return false;
    }
    try {
      const response = await this.getCurrentUser();
      currentUser = response.data;
      return true;
    } catch (e) {
      this.logout();
      return false;
    }
  },
};

// UI Helpers
const ui = {
  showAlert(message, type = 'success') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.textContent = message;
    document.body.insertBefore(alertDiv, document.body.firstChild);
    setTimeout(() => alertDiv.remove(), 4000);
  },

  showLoading(element, show = true) {
    if (show) {
      element.innerHTML = '<div class="spinner"></div>';
    }
  },

  formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-MY', { year: 'numeric', month: 'short', day: 'numeric' });
  },

  formatTime(timeStr) {
    const [hours, minutes] = timeStr.split(':');
    return `${hours}:${minutes}`;
  },

  getCurrentDate() {
    return new Date().toISOString().split('T')[0];
  },

  getWeekStart(date) {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1);
    return new Date(d.setDate(diff)).toISOString().split('T')[0];
  },

  getMonthStart(date) {
    const d = new Date(date);
    return new Date(d.getFullYear(), d.getMonth(), 1).toISOString().split('T')[0];
  },

  getMonthEnd(date) {
    const d = new Date(date);
    return new Date(d.getFullYear(), d.getMonth() + 1, 0).toISOString().split('T')[0];
  },
};

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
  if (api.isAuthenticated()) {
    const authenticated = await api.ensureAuthenticated();
    if (!authenticated) {
      window.location.href = '/login.html';
    }
  }
});
