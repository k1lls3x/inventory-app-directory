<template>
  <LoginForm v-if="!loggedIn" @login-success="onLoginSuccess" />
  <div v-else class="layout">
    <!-- Sidebar -->
    <aside class="sidebar">
  <!-- User Block -->
  <div class="sidebar-user" @click="showProfileModal = true" style="cursor:pointer;">
  <div class="sidebar-user-avatar">
    <svg width="40" height="40" fill="none" viewBox="0 0 40 40">
      <circle cx="20" cy="20" r="20" fill="#e3eaff"/>
      <path d="M20 24c-4 0-7 3-7 7h14c0-4-3-7-7-7Zm0-2a5 5 0 1 0 0-10 5 5 0 0 0 0 10Z" fill="#b7c5ec"/>
    </svg>
  </div>
  <div>
    <div class="sidebar-user-name">{{ user?.full_name || user?.username }}</div>
    <div class="sidebar-user-role">{{ roleName(user.role) }}</div>
  </div>
</div>

  <div class="sidebar-logo">
    <!-- SVG или логотип -->
  </div>
  <nav>
  <button
    v-for="tab in tabs"
    :key="tab"
    :class="{ active: currentTab === tab }"
    @click="currentTab = tab"
  >{{ tab }}</button>
</nav>
  <button class="logout-btn" @click="logout">🚪 Выйти</button>
</aside>

    <!-- Main Content -->
    <div class="main-content">
      <header class="main-header">
        <h1>Складская система</h1>
        <span class="username">Добро пожаловать!</span>
      </header>

      <main>
        <!-- Дашборд -->
        <section v-if="currentTab === 'Дашборд'">
          <div class="cards">
            <div class="card highlight animate-card">
              <p class="title">Всего остатков</p>
              <p class="value">{{ totalStock }}</p>
              <p class="note positive">+15% за месяц</p>
            </div>
            <div class="card animate-card">
              <p class="title">Товаров</p>
              <p class="value">{{ itemCount }}</p>
              <p class="note" v-if="newItems > 0">+{{ newItems }} новых за месяц</p>
            </div>
            <div class="card animate-card">
              <p class="title">Поставки</p>
              <p class="value">{{ monthlyOrders }}</p>
              <p class="note">в этом месяце</p>
            </div>
          </div>
          <div class="charts-table-wrap">
            <div class="chart-card animate-chart">
              <p class="title">Остатки за неделю</p>
              <LineChart v-if="weeklyStockChartData.datasets[0].data.length" :data="weeklyStockChartData" />
            </div>
            <div class="chart-card animate-chart">
              <p class="title">Оборот по складам</p>
              <LineChart v-if="turnoverLineChartData.datasets[0].data.length" :data="turnoverLineChartData" />
            </div>
          </div>
          <div class="table-section">
            <p class="title">Популярные товары</p>
            <table>
              <thead>
                <tr>
                  <th>Наименование</th>
                  <th>SKU</th>
                  <th>Склад</th>
                  <th>Остаток</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in topItems"
                  :key="item.id"
                  :class="{ 'zero-stock': item.quantity === 0 }"
                >
                  <td>{{ item.name }}</td>
                  <td>{{ item.sku }}</td>
                  <td>{{ item.warehouse }}</td>
                  <td>{{ item.quantity }}</td>
                </tr>
              </tbody>
            </table>
            <div v-if="topItems.length === 0" class="empty-message">
              Нет данных для отображения
            </div>
          </div>
          <div class="dashboard-risk-grid">
            <div class="table-section dashboard-risk-card">
              <div class="table-header dashboard-risk-header">
                <div>
                  <p class="title">Контроль дефицита</p>
                  <p class="dashboard-subtitle">Товары ниже минимального остатка и бюджет на пополнение</p>
                </div>
                <select v-model.number="selectedRiskWarehouseId" class="input dashboard-risk-filter">
                  <option :value="0">Все склады</option>
                  <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                    {{ wh.name }}
                  </option>
                </select>
              </div>

              <div class="risk-summary-cards">
                <div class="risk-summary-card warning">
                  <span class="risk-summary-label">Рисковых SKU</span>
                  <strong>{{ lowStockItems.length }}</strong>
                </div>
                <div class="risk-summary-card danger">
                  <span class="risk-summary-label">Критичных позиций</span>
                  <strong>{{ criticalLowStockCount }}</strong>
                </div>
                <div class="risk-summary-card neutral">
                  <span class="risk-summary-label">К дозакупке</span>
                  <strong>{{ recommendedRestockUnits }}</strong>
                </div>
                <div class="risk-summary-card accent">
                  <span class="risk-summary-label">Оценка бюджета</span>
                  <strong>{{ formatMoney(recommendedRestockCost) }}</strong>
                </div>
              </div>

              <table>
                <thead>
                  <tr>
                    <th>Товар</th>
                    <th>Склад</th>
                    <th>Остаток</th>
                    <th>Мин. остаток</th>
                    <th>Рекомендация</th>
                    <th>Статус</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in lowStockItems.slice(0, 8)" :key="`${item.item_id}-${item.warehouse_id}`">
                    <td>
                      <div>{{ item.name }}</div>
                      <div class="dashboard-row-meta">{{ item.sku }}</div>
                    </td>
                    <td>{{ item.warehouse }}</td>
                    <td>{{ item.current_stock }}</td>
                    <td>{{ item.reorder_level }}</td>
                    <td>{{ item.recommended_order }}</td>
                    <td>
                      <span class="stock-health-badge" :class="item.severity">
                        {{ item.status_label }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>

              <div v-if="lowStockItems.length === 0" class="empty-message">
                Все позиции находятся выше минимального остатка
              </div>
            </div>

            <div class="chart-card dashboard-risk-sidecard">
              <p class="title">Склады под нагрузкой</p>
              <div class="warehouse-risk-list">
                <div v-for="item in warehouseRiskSummary" :key="item.warehouse_id" class="warehouse-risk-row">
                  <div>
                    <div class="warehouse-risk-name">{{ item.name }}</div>
                    <div class="dashboard-row-meta">{{ item.low_count }} SKU требуют внимания</div>
                  </div>
                  <div class="warehouse-risk-values">
                    <strong>{{ item.critical_count }}</strong>
                    <span>критично</span>
                  </div>
                </div>
              </div>
              <div v-if="warehouseRiskSummary.length === 0" class="empty-message">
                По текущим данным дефицита нет
              </div>
            </div>
          </div>
        </section>
        <section v-if="currentTab === 'План закупок' && ['admin', 'manager'].includes(user?.role)">
          <div class="filters-bar">
            <div class="filter-group">
              <label>📦 Склад</label>
              <select v-model.number="selectedRiskWarehouseId" class="input">
                <option :value="0">Все склады</option>
                <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                  {{ wh.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>Поставщик</label>
              <select v-model.number="purchasePlanSupplierId" class="input">
                <option :value="0">Все поставщики</option>
                <option v-for="sup in suppliers" :key="sup.supplier_id" :value="sup.supplier_id">
                  {{ sup.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>Приоритет</label>
              <select v-model="purchasePlanPriority" class="input">
                <option value="all">Все позиции</option>
                <option value="order">К закупке</option>
                <option value="critical">Критично</option>
                <option value="warning">Низкий остаток</option>
                <option value="attention">Скоро закончится</option>
                <option value="stable">В норме</option>
              </select>
            </div>
            <div class="filter-group">
              <label>Горизонт</label>
              <select v-model.number="purchasePlanHorizon" class="input">
                <option :value="7">7 дней</option>
                <option :value="14">14 дней</option>
                <option :value="30">30 дней</option>
                <option :value="60">60 дней</option>
              </select>
            </div>
            <div class="filter-group">
              <label>🔍 Поиск</label>
              <input v-model="purchasePlanSearch" class="input" placeholder="Название, SKU или склад" />
            </div>
            <div class="filter-group button-group">
              <label>&nbsp;</label>
              <button class="add-button" @click="selectAllPurchaseRows">Выбрать к заказу</button>
            </div>
            <div class="filter-group button-group">
              <label>&nbsp;</label>
              <button class="export-button" @click="exportPurchasePlanCsv">CSV</button>
            </div>
          </div>

          <div class="cards">
            <div class="card animate-card">
              <p class="title">Позиций к закупке</p>
              <p class="value">{{ purchasePlanOrderCount }}</p>
              <p class="note">{{ purchasePlanAttentionCount }} под контролем</p>
            </div>
            <div class="card animate-card">
              <p class="title">Выбрано</p>
              <p class="value">{{ selectedPurchasePlanItems.length }}</p>
              <p class="note">строк в заявке</p>
            </div>
            <div class="card animate-card">
              <p class="title">Единиц к заказу</p>
              <p class="value">{{ selectedPurchasePlanUnits || filteredPurchasePlanUnits }}</p>
              <p class="note">по текущему фильтру</p>
            </div>
            <div class="card animate-card">
              <p class="title">Плановый бюджет</p>
              <p class="value">{{ formatMoney(selectedPurchasePlanCost || filteredPurchasePlanCost) }}</p>
            </div>
          </div>

          <div class="plan-workbench">
            <div class="plan-workbench-main">
              <div class="table-header">
                <div>
                  <p class="title">Рабочая заявка</p>
                  <p class="dashboard-subtitle">{{ purchasePlanOrderCount }} позиций требуют закупки на горизонте {{ purchasePlanHorizon }} дней</p>
                </div>
                <div class="plan-actions">
                  <button class="export-button" @click="copyPurchasePlan">Скопировать</button>
                  <button class="export-button" @click="clearPurchaseSelection">Очистить выбор</button>
                </div>
              </div>
              <div class="plan-chip-list">
                <button
                  v-for="item in purchasePrioritySummary"
                  :key="item.key"
                  class="plan-chip"
                  :class="{ active: purchasePlanPriority === item.key }"
                  @click="purchasePlanPriority = item.key"
                >
                  <span>{{ item.label }}</span>
                  <strong>{{ item.count }}</strong>
                </button>
              </div>
            </div>
            <div class="plan-workbench-side">
              <span class="risk-summary-label">Самый срочный товар</span>
              <strong>{{ topPurchaseRisk?.name || 'Нет срочных позиций' }}</strong>
              <span>{{ topPurchaseRisk ? `${topPurchaseRisk.warehouse}, ${topPurchaseRisk.days_left_label}` : 'Все позиции в норме' }}</span>
            </div>
          </div>

          <div class="table-section animate-table">
            <div class="table-header">
              <p class="title">План пополнения склада</p>
              <button
                class="add-button"
                :disabled="!selectedPurchasePlanItems.length"
                @click="startInboundFromPlan(selectedPurchasePlanItems[0])"
              >
                Создать поставку
              </button>
            </div>
            <table>
              <thead>
                <tr>
                  <th></th>
                  <th>Приоритет</th>
                  <th>Товар</th>
                  <th>SKU</th>
                  <th>Поставщик</th>
                  <th>Склад</th>
                  <th>Остаток</th>
                  <th>Мин. остаток</th>
                  <th>Расход/день</th>
                  <th>Хватит на</th>
                  <th>Заказать</th>
                  <th>Бюджет</th>
                  <th>Действие</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in filteredPurchasePlanItems" :key="`${item.item_id}-${item.warehouse_id}`">
                  <td>
                    <input
                      type="checkbox"
                      :checked="isPurchaseRowSelected(item)"
                      :disabled="item.recommended_order <= 0"
                      @change="togglePurchaseRow(item)"
                    />
                  </td>
                  <td>
                    <span class="stock-health-badge" :class="item.severity">
                      {{ item.status_label }}
                    </span>
                  </td>
                  <td>{{ item.name }}</td>
                  <td>{{ item.sku }}</td>
                  <td>{{ item.supplier_name || '—' }}</td>
                  <td>{{ item.warehouse }}</td>
                  <td>{{ item.current_stock }}</td>
                  <td>{{ item.reorder_level }}</td>
                  <td>{{ formatNumber(item.daily_demand, 1) }}</td>
                  <td>{{ item.days_left_label }}</td>
                  <td>{{ item.recommended_order }}</td>
                  <td>{{ formatMoney(item.estimated_cost) }}</td>
                  <td>
                    <button class="action-btn edit" @click="startInboundFromPlan(item)">➕</button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-if="filteredPurchasePlanItems.length === 0" class="empty-message">
              Нет позиций по текущим фильтрам
            </div>
          </div>
        </section>
 <!-- Пользователи (видно только админу) -->
 <section v-if="currentTab === 'Пользователи' && user?.role === 'admin'">
  <div class="filters-bar">
    <div class="filter-group">
      <label>🔍 Поиск</label>
      <input
        type="text"
        class="input"
        v-model="userSearch"
        placeholder="Имя пользователя, ФИО или роль"
      />
    </div>
    <div class="filter-group">
      <label>Роль</label>
      <select v-model="userRoleFilter" class="input">
        <option value="">Все роли</option>
        <option value="admin">Администраторы</option>
        <option value="manager">Менеджеры</option>
        <option value="worker">Сотрудники</option>
      </select>
    </div>
    <div class="filter-group button-group">
      <label>&nbsp;</label>
      <button class="add-button" @click="openAddUserModal">➕ Добавить пользователя</button>
    </div>
  </div>

  <div class="cards">
    <div class="card animate-card">
      <p class="title">Всего пользователей</p>
      <p class="value">{{ users.length }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Администраторы</p>
      <p class="value">{{ userRoleCounts.admin }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Менеджеры</p>
      <p class="value">{{ userRoleCounts.manager }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Сотрудники</p>
      <p class="value">{{ userRoleCounts.worker }}</p>
    </div>
  </div>

  <div class="table-section animate-table">
    <div class="table-header">
      <p class="title">Пользователи</p>
      <button class="export-button" @click="exportUsersToExcel">📤 Экспорт в Excel</button>
    </div>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Логин</th>
          <th>ФИО</th>
          <th>Роль</th>
          <th>Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in filteredUsers" :key="u.user_id">
          <td>{{ u.user_id }}</td>
          <td>{{ u.username }}</td>
          <td>{{ u.full_name }}</td>
          <td>{{ roleName(u.role) }}</td>
          <td>
            <div class="action-buttons">
              <button class="action-btn edit" @click="openEditUserModal(u)">✏️</button>
              <button class="action-btn delete" @click="deleteUser(u)">🗑️</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="filteredUsers.length === 0" class="empty-message">
      Нет пользователей по фильтру
    </div>
  </div>

  <!-- Модалка добавления пользователя -->
  <div v-if="showAddUserModal" class="modal-overlay" @click.self="closeAddUserModal">
    <div class="modal">
      <h3>Добавить пользователя</h3>
      <div class="form-group"><label>Логин</label><input v-model="newUser.username" /></div>
      <div class="form-group"><label>ФИО</label><input v-model="newUser.full_name" /></div>
      <div class="form-group"><label>Пароль</label><input type="password" v-model="newUser.password" /></div>
      <div class="form-group"><label>Роль</label>
        <select v-model="newUser.role">
          <option disabled value="">Выберите роль</option>
          <option value="admin">Администратор</option>
          <option value="manager">Менеджер</option>
          <option value="worker">Сотрудник</option>
        </select>
      </div>
      <div class="modal-actions">
        <button @click="confirmAddUser">💾 Сохранить</button>
        <button @click="closeAddUserModal">❌ Отмена</button>
      </div>
    </div>
  </div>

  <!-- Модалка редактирования пользователя -->
  <div v-if="showEditUserModal" class="modal-overlay" @click.self="closeEditUserModal">
    <div class="modal">
      <h3>Редактировать пользователя</h3>
      <div class="form-group"><label>Логин</label><input v-model="userToEdit.username" /></div>
      <div class="form-group"><label>ФИО</label><input v-model="userToEdit.full_name" /></div>
      <div class="form-group"><label>Новый пароль</label><input type="password" v-model="userToEdit.newPassword" placeholder="Оставьте пустым для без изменений" /></div>
      <div class="form-group"><label>Роль</label>
        <select v-model="userToEdit.role">
          <option value="admin">Администратор</option>
          <option value="manager">Менеджер</option>
          <option value="worker">Сотрудник</option>
        </select>
      </div>
      <div class="modal-actions">
        <button @click="confirmEditUser">💾 Сохранить</button>
        <button @click="closeEditUserModal">❌ Отмена</button>
      </div>
    </div>
  </div>
</section>

        <!-- Остатки -->
        <section v-if="currentTab === 'Остатки' && ['admin', 'manager', 'worker'].includes(user?.role)">
          <div class="filters-bar">
            <div class="filter-group">
              <label>📦 Склад</label>
              <select v-model="selectedWarehouseId" class="input">
                <option value="0">Все склады</option>
                <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                  {{ wh.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>🔍 Поиск</label>
              <input type="text" class="input" v-model="searchQuery" placeholder="Название, SKU или склад" />
            </div>
            <div class="filter-group">
              <label>Статус</label>
              <select v-model="stockStatusFilter" class="input">
                <option value="all">Все остатки</option>
                <option value="zero">Нулевые</option>
                <option value="low">Ниже минимума</option>
                <option value="watch">Под контролем</option>
                <option value="ok">В норме</option>
              </select>
            </div>
            <div class="filter-group button-group">
              <label>&nbsp;</label>
              <button class="add-button" @click="openAddModal">➕ Добавить остаток</button>
            </div>
          </div>

          <div class="cards">
            <div class="card animate-card">
              <p class="title">Строк остатков</p>
              <p class="value">{{ filteredStockList.length }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Единиц в фильтре</p>
              <p class="value">{{ filteredStockUnits }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Ниже минимума</p>
              <p class="value">{{ lowStockRowCount }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Нулевые позиции</p>
              <p class="value">{{ zeroStockRowCount }}</p>
            </div>
          </div>

          <!-- Модалка добавления остатка -->
          <div v-if="showAddModal" class="modal-overlay" @click.self="closeAddModal">
            <div class="modal">
              <h3>Добавить остаток</h3>
              <div class="form-group">
                <label for="item">Товар</label>
                <select v-model.number="newStock.item_id">
                  <option disabled value="0">Выберите товар</option>
                  <option v-for="item in items" :key="item.item_id" :value="item.item_id">
                    {{ item.name }} ({{ item.sku }})
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="warehouse">Склад</label>
                <select v-model.number="newStock.warehouse_id">
                  <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                    {{ wh.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="quantity">Количество</label>
                <input type="number" v-model.number="newStock.quantity" min="1" />
              </div>
              <div class="modal-actions">
                <button @click="confirmAddStock">💾 Сохранить</button>
                <button @click="closeAddModal">❌ Отмена</button>
              </div>
            </div>
          </div>

          <!-- Модалка редактирования остатка -->
          <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
            <div class="modal">
              <h3>Редактировать остаток</h3>
              <div class="form-group">
                <label for="item">Товар</label>
                <input type="text" :value="stockToEdit?.name" disabled />
              </div>
              <div class="form-group">
                <label for="warehouse">Склад</label>
                <input type="text" :value="stockToEdit?.warehouse" disabled />
              </div>
              <div class="form-group">
                <label for="quantity">Количество</label>
                <input type="number" v-model.number="stockToEdit.quantity" min="1" />
              </div>
              <div class="modal-actions">
                <button @click="confirmEditStock">💾 Сохранить</button>
                <button @click="closeEditModal">❌ Отмена</button>
              </div>
            </div>
          </div>

          <div class="charts-table-wrap">
            <div class="chart-card animate-chart">
              <BarChart v-if="filteredChartData.datasets[0].data.length" :data="filteredChartData" />
            </div>
            <div class="table-section animate-table">
              <div class="table-header">
                <p class="title">Остатки на складе</p>
                <button class="export-button" @click="exportToExcel">📤 Экспорт в Excel</button>
              </div>
              <table>
                <thead>
                  <tr>
                    <th>Наименование</th>
                    <th>Номер</th>
                    <th>Склад</th>
                    <th>Количество</th>
                    <th>Мин. остаток</th>
                    <th>Статус</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="stock in filteredStockList" :key="stock.id">
                    <td>{{ stock.name }}</td>
                    <td>{{ stock.sku }}</td>
                    <td>{{ stock.warehouse }}</td>
                    <td>{{ stock.quantity }}</td>
                    <td>{{ stock.reorder_level }}</td>
                    <td>
                      <span class="stock-health-badge" :class="stock.stock_status">
                        {{ stock.status_label }}
                      </span>
                    </td>
                    <td>
                      <div class="action-buttons">
                        <button class="action-btn edit" @click="openEditModal(stock)">✏️</button>
                        <button class="action-btn edit" @click="openTransferFromStock(stock)">↔</button>
                        <button class="action-btn delete" @click="deleteStock(stock)">🗑️</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
              <div v-if="filteredStockList.length === 0" class="empty-message">
                Нет остатков для отображения
              </div>
            </div>
          </div>
        </section>

        <!-- Поставки -->
        <section v-if="currentTab === 'Поставки' && ['admin', 'manager', 'worker'].includes(user?.role)">          <div class="filters-bar">
            <div class="filter-group">
              <label>📅 Дата</label>
              <input type="date" class="input" v-model="selectedDeliveryDate" :max="'3030-12-31'" />
            </div>
            <div class="filter-group">
              <label>Склад</label>
              <select v-model.number="selectedDeliveryWarehouse" class="input">
                <option :value="0">Все склады</option>
                <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                  {{ wh.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>Поставщик</label>
              <select v-model.number="selectedDeliverySupplier" class="input">
                <option :value="0">Все поставщики</option>
                <option v-for="sup in suppliers" :key="sup.supplier_id" :value="sup.supplier_id">
                  {{ sup.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>🔍 Поиск</label>
              <input type="text" class="input" v-model="deliverySearchQuery" placeholder="Название, SKU или поставщик" />
            </div>
            <div class="filter-group button-group">
              <label>&nbsp;</label>
              <button class="add-button" @click="openAddDeliveryModal">➕ Добавить поставку</button>
            </div>
          </div>

          <div class="cards">
            <div class="card animate-card">
              <p class="title">Поставок в фильтре</p>
              <p class="value">{{ filteredDeliveriesList.length }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Принято единиц</p>
              <p class="value">{{ filteredDeliveryUnits }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Поставщиков</p>
              <p class="value">{{ filteredDeliverySupplierCount }}</p>
            </div>
            <div class="card animate-card">
              <p class="title">Средняя поставка</p>
              <p class="value">{{ formatNumber(averageDeliverySize, 0) }}</p>
            </div>
          </div>

          <!-- Модалка добавления поставки -->
          <div v-if="showAddDeliveryModal" class="modal-overlay" @click.self="closeAddDeliveryModal">
            <div class="modal">
              <h3>Добавить поставку</h3>
              <div class="form-group">
                <label for="inbound-item">Товар</label>
                <select v-model.number="newInbound.item_id" id="inbound-item">
                  <option disabled value="0">Выберите товар</option>
                  <option value="-1">➕ Новый товар...</option>
                  <option v-for="item in items" :key="item.item_id" :value="item.item_id">
                    {{ item.name }} ({{ item.sku }})
                  </option>
                </select>
                <div v-if="newInbound.item_id === -1" class="new-item-fields">
                  <input class="input" placeholder="SKU" v-model="newInboundItem.sku" />
                  <input class="input" placeholder="Наименование" v-model="newInboundItem.name" />
                  <input class="input" placeholder="Описание" v-model="newInboundItem.description" />
                  <input class="input" placeholder="Ед. изм." v-model="newInboundItem.uom" />
                  <input class="input" type="number" placeholder="Цена" v-model.number="newInboundItem.price" />
                  <input class="input" type="number" placeholder="Себестоимость" v-model.number="newInboundItem.cost" />
                </div>
              </div>
              <div class="form-group">
                <label for="inbound-supplier">От кого: поставщик</label>
                <select v-model.number="newInbound.supplier_id" id="inbound-supplier">
                  <option disabled value="0">Выберите поставщика</option>
                  <option v-for="sup in suppliers" :key="sup.supplier_id" :value="sup.supplier_id">
                    {{ sup.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="inbound-warehouse">Куда: склад-получатель</label>
                <select v-model.number="newInbound.warehouse_id" id="inbound-warehouse">
                  <option disabled value="0">Выберите склад</option>
                  <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                    {{ wh.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="inbound-quantity">Количество</label>
                <input type="number" min="1" v-model.number="newInbound.quantity" id="inbound-quantity" />
              </div>
              <div class="form-group">
                <label for="inbound-date">Дата поступления</label>
                <input type="date" v-model="newInbound.received_at" id="inbound-date" :max="'3030-12-31'" />
              </div>
              <div class="form-group">
                <label for="inbound-document">Накладная / акт</label>
                <input v-model="newInbound.document_no" id="inbound-document" placeholder="Например: ТОРГ-12 №154" />
              </div>
              <div class="form-group">
                <label for="inbound-receiver">Кто принял</label>
                <select v-model.number="newInbound.received_by" id="inbound-receiver">
                  <option disabled value="0">Выберите сотрудника</option>
                  <option v-for="u in responsibleUsers" :key="u.user_id" :value="u.user_id">
                    {{ u.full_name || u.username }} — {{ roleName(u.role) }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="inbound-note">Комментарий</label>
                <input v-model="newInbound.note" id="inbound-note" placeholder="Например: принято по доверенности" />
              </div>
              <div class="modal-actions">
                <button @click="confirmAddDelivery">💾 Сохранить</button>
                <button @click="closeAddDeliveryModal">❌ Отмена</button>
              </div>
            </div>
          </div>

          <!-- Модалка редактирования поставки -->
          <div v-if="showEditDeliveryModal && deliveryToEdit" class="modal-overlay" @click.self="closeEditDeliveryModal">
            <div class="modal">
              <h3>Редактировать поставку</h3>
              <div class="form-group">
                <label for="edit-inbound-item">Товар</label>
                <select v-model.number="deliveryToEdit.item_id" id="edit-inbound-item">
                  <option disabled value="0">Выберите товар</option>
                  <option v-for="item in items" :key="item.item_id" :value="item.item_id">
                    {{ item.name }} ({{ item.sku }})
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="edit-inbound-supplier">От кого: поставщик</label>
                <select v-model.number="deliveryToEdit.supplier_id" id="edit-inbound-supplier">
                  <option disabled value="0">Выберите поставщика</option>
                  <option v-for="sup in suppliers" :key="sup.supplier_id" :value="sup.supplier_id">
                    {{ sup.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="edit-inbound-warehouse">Куда: склад-получатель</label>
                <select v-model.number="deliveryToEdit.warehouse_id" id="edit-inbound-warehouse">
                  <option disabled value="0">Выберите склад</option>
                  <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
                    {{ wh.name }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="edit-inbound-quantity">Количество</label>
                <input type="number" min="1" v-model.number="deliveryToEdit.quantity" id="edit-inbound-quantity" />
              </div>
              <div class="form-group">
                <label for="edit-inbound-date">Дата поступления</label>
                <input type="date" v-model="deliveryToEdit.received_at" id="edit-inbound-date" :max="'3030-12-31'" />
              </div>
              <div class="form-group">
                <label for="edit-inbound-document">Накладная / акт</label>
                <input v-model="deliveryToEdit.document_no" id="edit-inbound-document" />
              </div>
              <div class="form-group">
                <label for="edit-inbound-receiver">Кто принял</label>
                <select v-model.number="deliveryToEdit.received_by" id="edit-inbound-receiver">
                  <option disabled value="0">Выберите сотрудника</option>
                  <option v-for="u in responsibleUsers" :key="u.user_id" :value="u.user_id">
                    {{ u.full_name || u.username }} — {{ roleName(u.role) }}
                  </option>
                </select>
              </div>
              <div class="form-group">
                <label for="edit-inbound-note">Комментарий</label>
                <input v-model="deliveryToEdit.note" id="edit-inbound-note" />
              </div>
              <div class="modal-actions">
                <button @click="confirmEditDelivery">💾 Сохранить</button>
                <button @click="closeEditDeliveryModal">❌ Отмена</button>
              </div>
            </div>
          </div>

          <div class="charts-table-wrap">
            <div class="chart-card animate-chart">
              <BarChart v-if="filteredDeliveriesChartData.datasets[0].data.length" :data="filteredDeliveriesChartData" />
            </div>
            <div class="table-section animate-table">
              <div class="table-header">
                <p class="title">Поставки</p>
                <button class="export-button" @click="exportDeliveriesToExcel">📤 Экспорт в Excel</button>
              </div>
              <table>
                <thead>
                  <tr>
                    <th>Дата</th>
                    <th>Документ</th>
                    <th>Наименование</th>
                    <th>SKU</th>
                    <th>Куда</th>
                    <th>От кого</th>
                    <th>Принял</th>
                    <th>Количество</th>
                    <th>Комментарий</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="d in filteredDeliveriesList" :key="d.inbound_id">
                    <td>{{ formatDate(d.date) }}</td>
                    <td>{{ d.document_no || '—' }}</td>
                    <td>{{ d.name }}</td>
                    <td>{{ d.sku }}</td>
                    <td>{{ d.warehouse }}</td>
                    <td>{{ d.supplier }}</td>
                    <td>{{ d.receiver_name || '—' }}</td>
                    <td>{{ d.quantity }}</td>
                    <td>{{ d.note || '—' }}</td>
                    <td>
                      <div class="action-buttons">
                        <button class="action-btn edit" @click="openEditDeliveryModal(d)">✏️</button>
                        <button class="action-btn delete" @click="deleteDelivery(d)">🗑️</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
              <div v-if="filteredDeliveriesList.length === 0" class="empty-message">
                Нет поставок за выбранную дату
              </div>
            </div>
          </div>

        </section>

        <!-- Товары -->
        <section v-if="currentTab === 'Товары' && ['admin', 'manager'].includes(user?.role)">
          <div class="filters-bar">
        <div class="filter-group">
          <label>🔍 Поиск</label>
          <input type="text" class="input" v-model="itemSearch" placeholder="Название или SKU" />
        </div>
        <div class="filter-group">
          <label>Статус</label>
          <select v-model="itemStatusFilter" class="input">
            <option value="all">Все товары</option>
            <option value="no_stock">Без остатка</option>
            <option value="low">Ниже минимума</option>
            <option value="profitable">С маржой</option>
            <option value="no_price">Без цены</option>
          </select>
        </div>
        <div class="filter-group button-group">
          <label>&nbsp;</label>
          <button class="add-button" @click="openAddItemModal">➕ Добавить товар</button>
        </div>
      </div>
      <div class="cards">
          <div class="card animate-card">
            <p class="title">Всего товаров</p>
            <p class="value">{{ items.length }}</p>
          </div>
          <div class="card animate-card">
            <p class="title">Средняя цена товара</p>
            <p class="value">{{ averagePrice }}</p>
          </div>
          <div class="card animate-card">
            <p class="title">Наибольший остаток</p>
            <p class="value">
              {{ maxActualStock }}
            </p>
          </div>
          <div class="card animate-card">
            <p class="title">Средняя маржа</p>
            <p class="value">{{ averageMarginPercent }}%</p>
          </div>
        </div>

          <div class="table-section animate-table">
            <div class="table-header">
              <p class="title">Товары</p>
              <button class="export-button" @click="exportItemsToExcel">📤 Экспорт в Excel</button>
            </div>
            <table>
                  <thead>
                    <tr>
                      <th>Наименование</th>
                      <th>SKU</th>
                      <th>Ед. изм.</th>
                      <th>Описание</th>
                      <th>Остаток</th>
                      <th>Мин.</th>
                      <th>Партия</th>
                      <th>Цена</th>
                      <th>Себестоимость</th>
                      <th>Маржа</th>
                      <th>Действия</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in filteredItems" :key="item.item_id">
                      <td>{{ item.name }}</td>
                      <td>{{ item.sku }}</td>
                      <td>{{ item.uom }}</td>
                      <td>{{ item.description }}</td>
                      <td>{{ itemStockTotal(item.item_id) }}</td>
                      <td>{{ item.reorder_level || 0 }}</td>
                      <td>{{ item.reorder_qty || 0 }}</td>
                      <td>
                        {{ formatMoneyValue(itemPrice(item)) }}
                      </td>
                      <td>{{ formatMoneyValue(itemCost(item)) }}</td>
                      <td>{{ itemMarginPercent(item) }}%</td>
                      <td>
                        <div class="action-buttons">
                          <button class="action-btn edit" @click="openEditItemModal(item)">✏️</button>
                          <button class="action-btn delete" @click="deleteItem(item)">🗑️</button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>

            <div v-if="filteredItems.length === 0" class="empty-message">
              Нет товаров по фильтру
            </div>
          </div>
          <div v-if="showAddItemModal" class="modal-overlay" @click.self="showAddItemModal = false">
  <div class="modal">
    <h3>Добавить товар</h3>
    <div class="form-group"><label>SKU</label><input v-model="newItem.sku" /></div>
<div class="form-group"><label>Наименование</label><input v-model="newItem.name" /></div>
<div class="form-group"><label>Описание</label><input v-model="newItem.description" /></div>
<div class="form-group"><label>Ед. изм.</label><input v-model="newItem.uom" /></div>
<div class="form-row">
  <div class="form-group half"><label>Мин. остаток</label><input type="number" min="0" v-model.number="newItem.reorder_level" /></div>
  <div class="form-group half"><label>Партия для дозакупки</label><input type="number" min="0" v-model.number="newItem.reorder_qty" /></div>
</div>
<div class="form-group"><label>Цена</label><input type="number" v-model.number="newItem.price" /></div>
<div class="form-group"><label>Себестоимость</label><input type="number" v-model.number="newItem.cost" /></div>

    <div class="modal-actions">
      <button @click="confirmAddItem">💾 Сохранить</button>
      <button @click="showAddItemModal = false">❌ Отмена</button>
    </div>
  </div>
</div>
<div v-if="showEditItemModal" class="modal-overlay" @click.self="showEditItemModal = false">
  <div class="modal item-edit-modal">
    <h3>Редактировать товар</h3>
    <form @submit.prevent="confirmEditItem" autocomplete="off">
      <div class="form-group"><label>SKU</label>
        <input v-model="itemToEdit.sku" disabled class="input-modern" />
      </div>
      <div class="form-group"><label>Наименование</label>
        <input v-model="itemToEdit.name" class="input-modern" />
      </div>
      <div class="form-group"><label>Описание</label>
        <textarea v-model="itemToEdit.description" rows="2" class="input-modern" style="resize:vertical; min-height:36px;" />
      </div>
      <div class="form-row">
        <div class="form-group half"><label>Категория</label>
          <input v-model="itemToEdit.category" class="input-modern" />
        </div>
        <div class="form-group half"><label>Ед. изм.</label>
          <input v-model="itemToEdit.uom" class="input-modern" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group half"><label>Мин. остаток</label>
          <input type="number" v-model.number="itemToEdit.reorder_level" min="0" class="input-modern" />
        </div>
        <div class="form-group half"><label>Партия для дозакупки</label>
          <input type="number" v-model.number="itemToEdit.reorder_qty" min="0" class="input-modern" />
        </div>
      </div>
      <div class="form-row">
        <div class="form-group half"><label>Цена</label>
          <input type="number" v-model.number="itemToEdit.price" min="0" step="0.01" class="input-modern" />
        </div>
        <div class="form-group half"><label>Себестоимость</label>
          <input type="number" v-model.number="itemToEdit.cost" min="0" step="0.01" class="input-modern" />
        </div>
      </div>
      <div class="modal-actions modal-actions-row">
        <button type="submit" class="main-btn-strong">💾 Сохранить</button>
        <button type="button" class="main-btn-ghost" @click="showEditItemModal = false">❌ Отмена</button>
      </div>
    </form>
  </div>
</div>



        </section>
        <section v-if="currentTab === 'Склады'">
  <!-- Карточки -->
  <div class="cards">
    <div class="card animate-card">
      <p class="title">Всего складов</p>
      <p class="value">{{ warehouses.length }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Складов с остатками</p>
      <p class="value">{{ activeWarehouseCount }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Единиц на складах</p>
      <p class="value">{{ totalStock }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Склады с рисками</p>
      <p class="value">{{ warehouseRiskSummary.length }}</p>
    </div>
  </div>

  <!-- Фильтры -->
  <div class="filters-bar">
    <div class="filter-group">
      <label>🔍 Поиск</label>
      <input
        type="text"
        class="input"
        v-model="warehouseSearch"
        placeholder="Название или локация"
      />
    </div>
    <div class="filter-group button-group">
      <label>&nbsp;</label>
      <button class="add-button" @click="openAddModal">➕ Добавить склад</button>
    </div>
  </div>

  <!-- Таблица -->
  <div class="table-section animate-table">
    <div class="table-header">
      <p class="title">Склады</p>
      <!-- <button class="export-button">📤 Экспорт</button> -->
    </div>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Название</th>
          <th>Локация</th>
          <th>SKU</th>
          <th>Остаток</th>
          <th>Риски</th>
          <th>Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="w in filteredWarehouses" :key="w.warehouse_id">
          <td>{{ w.warehouse_id }}</td>
          <td>{{ w.name }}</td>
          <td>{{ w.location }}</td>
          <td>{{ warehouseStats(w.warehouse_id).sku_count }}</td>
          <td>{{ warehouseStats(w.warehouse_id).units }}</td>
          <td>{{ warehouseStats(w.warehouse_id).risk_count }}</td>
          <td>
            <div class="action-buttons">
              <button class="action-btn edit" @click="editWarehouse(w)">✏️</button>
              <!-- Тут не делаем кнопку “Удалить”, т.к. удаление складов нежелательно -->
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="filteredWarehouses.length === 0" class="empty-message">
      Нет складов по фильтру
    </div>
  </div>

  <!-- Модалка добавления склада -->
  <div v-if="showAddModal" class="modal-overlay" @click.self="closeAddModal">
    <div class="modal">
      <h3>Добавить склад</h3>
      <div class="form-group">
        <label>Название склада</label>
        <input v-model="newWarehouse.name" />
      </div>
      <div class="form-group">
        <label>Локация</label>
        <input v-model="newWarehouse.location" />
      </div>
      <div class="modal-actions">
        <button @click="addWarehouse">💾 Сохранить</button>
        <button @click="closeAddModal">❌ Отмена</button>
      </div>
    </div>
  </div>

  <!-- Модалка редактирования склада -->
  <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
    <div class="modal">
      <h3>Редактировать склад</h3>
      <div class="form-group">
        <label>Название склада</label>
        <input v-model="editWarehouseData.name" />
      </div>
      <div class="form-group">
        <label>Локация</label>
        <input v-model="editWarehouseData.location" />
      </div>
      <div class="modal-actions">
        <button @click="updateWarehouse">💾 Сохранить</button>
        <button @click="closeEditModal">❌ Отмена</button>
      </div>
    </div>
  </div>
</section>
<section v-if="currentTab === 'Движения'">
  <div class="filters-bar">
    <div class="filter-group">
      <label>Тип движения</label>
      <select v-model="moveType" class="input">
        <option value="">Все</option>
        <option value="inbound">Поступление</option>
        <option value="outbound">Отгрузка</option>
        <option value="transfer">Перемещение</option>
      </select>
    </div>
    <div class="filter-group">
      <label>📦 Склад</label>
      <select v-model.number="moveWarehouseId" class="input">
        <option value="0">Все склады</option>
        <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">{{ wh.name }}</option>
      </select>
    </div>
    <div class="filter-group">
      <label>С даты</label>
      <input type="date" v-model="movementDateFrom" class="input" />
    </div>
    <div class="filter-group">
      <label>По дату</label>
      <input type="date" v-model="movementDateTo" class="input" />
    </div>
    <div class="filter-group">
      <label>🔍 Поиск товара</label>
      <input v-model="moveItemSearch" placeholder="Название или SKU" class="input" />
    </div>
    <div class="filter-group button-group">
      <label>&nbsp;</label>
      <button class="add-button" @click="openTransferModal">➕ Переместить товар</button>
    </div>
    <div class="filter-group button-group">
      <label>&nbsp;</label>
      <button class="export-button" @click="exportMovementsCsv">CSV</button>
    </div>
  </div>

  <div class="cards">
    <div class="card animate-card">
      <p class="title">Всего поступлений</p>
      <p class="value">{{ inboundCount }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Всего отгрузок</p>
      <p class="value">{{ outboundCount }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Общий оборот</p>
      <p class="value">{{ totalMoved }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Перемещений</p>
      <p class="value">{{ transferCount }}</p>
    </div>
  </div>

  <div class="charts-table-wrap">
    <div class="chart-card animate-chart">
      <LineChart :data="moveChartData" />
    </div>
  </div>

  <div class="table-section animate-table">
    <div class="table-header">
      <p class="title">Движения товаров</p>
    </div>
    <table>
      <thead>
        <tr>
          <th>Дата</th>
          <th>Тип</th>
          <th>Товар</th>
          <th>Склад</th>
          <th>Кол-во</th>
          <th>Поставщик</th>
          <th>Описание</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="m in filteredMovements" :key="m.movement_id">
  <td>{{ formatDate(m.date) }}</td>
  <td>{{ movementTypeName(m.type) }}</td>
  <td>{{ m.item_name }}</td>
  <td>{{ m.warehouse_name }}</td>
  <td :class="{ 'positive': m.type==='inbound','negative': m.type==='outbound' }">
    {{ m.quantity }}
  </td>
  <!-- всегда рендерим две ячейки в одном порядке -->
  <td>
    {{ m.type === 'inbound' ? (m.supplier_name || '—') : '—' }}
  </td>
  <td>
    {{ ['outbound', 'transfer'].includes(m.type) ? (m.destination || '—') : '—' }}
  </td>
</tr>
      </tbody>
    </table>
    <div v-if="filteredMovements.length === 0" class="empty-message">Нет движений по фильтру</div>
  </div>

  <div class="table-section animate-table">
    <div class="table-header">
      <p class="title">Межскладские перемещения</p>
    </div>
    <table>
      <thead>
        <tr>
          <th>Дата</th>
          <th>Товар</th>
          <th>SKU</th>
          <th>Со склада</th>
          <th>На склад</th>
          <th>Кол-во</th>
          <th>Комментарий</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in filteredTransferList" :key="t.transfer_id">
          <td>{{ formatDate(t.date) }}</td>
          <td>{{ t.item_name }}</td>
          <td>{{ t.sku }}</td>
          <td>{{ t.from_warehouse }}</td>
          <td>{{ t.to_warehouse }}</td>
          <td>{{ t.quantity }}</td>
          <td>{{ t.note || '—' }}</td>
        </tr>
      </tbody>
    </table>
    <div v-if="filteredTransferList.length === 0" class="empty-message">Перемещений по фильтру нет</div>
  </div>

  <div v-if="showTransferModal" class="modal-overlay" @click.self="closeTransferModal">
    <div class="modal">
      <h3>Переместить товар</h3>
      <div class="form-group">
        <label>Товар</label>
        <select v-model.number="newTransfer.item_id">
          <option disabled value="0">Выберите товар</option>
          <option v-for="item in items" :key="item.item_id" :value="item.item_id">
            {{ item.name }} ({{ item.sku }})
          </option>
        </select>
      </div>
      <div class="form-row">
        <div class="form-group half">
          <label>Со склада</label>
          <select v-model.number="newTransfer.from_warehouse_id">
            <option disabled value="0">Выберите склад</option>
            <option v-for="w in warehouses" :key="w.warehouse_id" :value="w.warehouse_id">
              {{ w.name }}
            </option>
          </select>
        </div>
        <div class="form-group half">
          <label>На склад</label>
          <select v-model.number="newTransfer.to_warehouse_id">
            <option disabled value="0">Выберите склад</option>
            <option v-for="w in warehouses" :key="w.warehouse_id" :value="w.warehouse_id">
              {{ w.name }}
            </option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <label>Доступно на складе-источнике</label>
        <input :value="transferAvailableStock" disabled />
      </div>
      <div class="form-group">
        <label>Количество</label>
        <input type="number" min="1" v-model.number="newTransfer.quantity" />
      </div>
      <div class="form-group">
        <label>Дата перемещения</label>
        <input type="date" v-model="newTransfer.transferred_at" />
      </div>
      <div class="form-group">
        <label>Комментарий</label>
        <input v-model="newTransfer.note" placeholder="Например: пополнение зоны выдачи" />
      </div>
      <div class="modal-actions">
        <button @click="confirmTransfer">💾 Провести</button>
        <button @click="closeTransferModal">❌ Отмена</button>
      </div>
    </div>
  </div>
</section>

<!-- Поставщики -->
<section v-if="currentTab === 'Поставщики' && ['admin', 'manager'].includes(user?.role)">


<div class="filters-bar">
  <div class="filter-group">
    <label>🔍 Поиск</label>
    <input type="text" class="input" v-model="supplierSearch" placeholder="Название или ИНН" />
  </div>
  <div class="filter-group">
    <label>Активность</label>
    <select v-model="supplierActivityFilter" class="input">
      <option value="all">Все поставщики</option>
      <option value="active">С поставками</option>
      <option value="idle">Без поставок</option>
    </select>
  </div>
  <div class="filter-group button-group">
    <label>&nbsp;</label>
    <button class="add-button" @click="openAddSupplierModal">➕ Добавить поставщика</button>
  </div>
</div>

<div class="cards">
  <div class="card animate-card">
    <p class="title">Всего поставщиков</p>
    <p class="value">{{ suppliers.length }}</p>
  </div>
  <div class="card animate-card">
    <p class="title">Активных</p>
    <p class="value">{{ activeSupplierCount }}</p>
  </div>
  <div class="card animate-card">
    <p class="title">Принято единиц</p>
    <p class="value">{{ supplierInboundUnits }}</p>
  </div>
  <div class="card animate-card">
    <p class="title">Поставок</p>
    <p class="value">{{ deliveriesList.length }}</p>
  </div>
</div>

<div class="table-section animate-table">
  <div class="table-header">
    <p class="title">Поставщики</p>
    <button class="export-button" @click="exportSuppliersToExcel">📤 Экспорт в Excel</button>
  </div>
  <table>
    <thead>
      <tr>
        <th>Название</th>
        <th>ИНН</th>
        <th>Контакт</th>
        <th>Телефон</th>
        <th>Email</th>
        <th>Поставки</th>
        <th>Единиц</th>
        <th>Последняя</th>
        <th>Действия</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="s in filteredSuppliers" :key="s.supplier_id">
        <td>{{ s.name }}</td>
        <td>{{ s.inn }}</td>
        <td>{{ s.contact_person }}</td>
        <td>{{ s.phone }}</td>
        <td>{{ s.email }}</td>
        <td>{{ supplierStats(s.supplier_id).deliveries }}</td>
        <td>{{ supplierStats(s.supplier_id).units }}</td>
        <td>{{ supplierStats(s.supplier_id).last_date ? formatDate(supplierStats(s.supplier_id).last_date) : '—' }}</td>
        <td>
          <div class="action-buttons">
            <button class="action-btn edit" @click="openEditSupplierModal(s)">✏️</button>
            <button class="action-btn delete" @click="deleteSupplier(s)">🗑️</button>
          </div>
        </td>
      </tr>
    </tbody>
  </table>
  <div v-if="filteredSuppliers.length === 0" class="empty-message">
    Нет поставщиков по фильтру
  </div>
</div>

<!-- Модалка добавления -->
<div v-if="showAddSupplierModal" class="modal-overlay" @click.self="showAddSupplierModal = false">
  <div class="modal">
    <h3>Добавить поставщика</h3>
    <div class="form-group"><label>Название</label><input v-model="newSupplier.name" /></div>
    <div class="form-group"><label>ИНН</label><input v-model="newSupplier.inn" /></div>
    <div class="form-group"><label>Контакт</label><input v-model="newSupplier.contact_person" /></div>
    <div class="form-group"><label>Телефон</label> <input
          v-model="newSupplier.phone"
          @input="maskPhone($event, newSupplier)"
          maxlength="18"
          placeholder="+7 (___)-___-__-__"
          type="tel"
        /></div>
    <div class="form-group"><label>Email</label><input v-model="newSupplier.email" /></div>
    <div class="modal-actions">
      <button @click="confirmAddSupplier">💾 Сохранить</button>
      <button @click="showAddSupplierModal = false">❌ Отмена</button>
    </div>
  </div>
</div>

<!-- Модалка редактирования -->
<div v-if="showEditSupplierModal" class="modal-overlay" @click.self="showEditSupplierModal = false">
  <div class="modal">
    <h3>Редактировать поставщика</h3>
    <div class="form-group"><label>Название</label><input v-model="supplierToEdit.name" /></div>
    <div class="form-group"><label>ИНН</label><input v-model="supplierToEdit.inn" /></div>
    <div class="form-group"><label>Контакт</label><input v-model="supplierToEdit.contact_person" /></div>
    <div class="form-group"><label>Телефон</label><input
  v-model="supplierToEdit.phone"
  @input="maskPhone($event, supplierToEdit)"
  maxlength="18"
  placeholder="+7 (___)-___-__-__"
  type="tel"
/></div>
    <div class="form-group"><label>Email</label><input v-model="supplierToEdit.email" /></div>
    <div class="modal-actions">
      <button @click="confirmEditSupplier">💾 Сохранить</button>
      <button @click="showEditSupplierModal = false">❌ Отмена</button>
    </div>
  </div>
</div>
</section>
<section v-if="currentTab === 'Отгрузки'">
  <div class="filters-bar">
    <div class="filter-group">
      <label>📅 Дата</label>
      <input type="date" v-model="outboundDateFilter" class="input" />
    </div>
    <div class="filter-group">
      <label>Склад</label>
      <select v-model.number="outboundWarehouseFilter" class="input">
        <option :value="0">Все склады</option>
        <option v-for="wh in warehouses" :key="wh.warehouse_id" :value="wh.warehouse_id">
          {{ wh.name }}
        </option>
      </select>
    </div>
    <div class="filter-group">
      <label>🔍 Поиск</label>
      <input v-model="outboundSearch" class="input" placeholder="Товар, SKU, склад, получатель" />
    </div>
    <div class="filter-group button-group">
      <label>&nbsp;</label>
      <button class="add-button" @click="openAddOutboundModal">➕ Добавить отгрузку</button>
    </div>
  </div>

  <div class="cards">
    <div class="card animate-card">
      <p class="title">Всего отгрузок</p>
      <p class="value">{{ filteredOutboundList.length }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Суммарно отгружено</p>
      <p class="value">{{ totalOutboundQuantity }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Получателей</p>
      <p class="value">{{ outboundDestinationCount }}</p>
    </div>
    <div class="card animate-card">
      <p class="title">Средняя отгрузка</p>
      <p class="value">{{ formatNumber(averageOutboundSize, 0) }}</p>
    </div>
  </div>

  <div class="charts-table-wrap">
    <div class="chart-card animate-chart">
      <LineChart :data="outboundChartData" />
    </div>
  </div>

  <div class="table-section animate-table">
    <div class="table-header">
      <p class="title">Отгрузки</p>
    </div>
    <table>
      <thead>
        <tr>
          <th>Дата</th>
          <th>Товар</th>
          <th>SKU</th>
          <th>Документ</th>
          <th>Со склада</th>
          <th>Кому / назначение</th>
          <th>Отпустил</th>
          <th>Кол-во</th>
          <th>Комментарий</th>
          <th>Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in filteredOutboundList" :key="o.outbound_id">
          <td>{{ formatDate(o.date) }}</td>
          <td>{{ o.name }}</td>
          <td>{{ o.sku }}</td>
          <td>{{ o.document_no || '—' }}</td>
          <td>{{ o.warehouse }}</td>
          <td>{{ o.destination }}</td>
          <td>{{ o.shipper_name || '—' }}</td>
          <td class="negative">{{ o.quantity }}</td>
          <td>{{ o.note || '—' }}</td>
          <td>
            <div class="action-buttons">
              <button class="action-btn edit" @click="openEditOutboundModal(o)">✏️</button>
              <button class="action-btn delete" @click="deleteOutbound(o)">🗑️</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="filteredOutboundList.length === 0" class="empty-message">
      Нет отгрузок по фильтру
    </div>
  </div>

  <!-- Модалка добавления -->
  <div v-if="showAddOutboundModal" class="modal-overlay" @click.self="closeAddOutboundModal">
    <div class="modal">
      <h3>Добавить отгрузку</h3>
      <div class="form-group"><label>Товар</label>
        <select v-model.number="newOutbound.item_id">
          <option disabled value="0">Выберите товар</option>
          <option v-for="item in items" :key="item.item_id" :value="item.item_id">{{ item.name }} ({{ item.sku }})</option>
        </select>
      </div>
      <div class="form-group"><label>Со склада</label>
        <select v-model.number="newOutbound.warehouse_id">
          <option disabled value="0">Выберите склад</option>
          <option v-for="w in warehouses" :key="w.warehouse_id" :value="w.warehouse_id">{{ w.name }}</option>
        </select>
      </div>
      <div class="form-group"><label>Кому / назначение</label>
        <input v-model="newOutbound.destination" placeholder="Получатель, цех, заказ или объект" />
      </div>
      <div class="form-group"><label>Количество</label>
        <input type="number" min="1" v-model.number="newOutbound.quantity" />
      </div>
      <div class="form-group">
        <label>Доступно на складе</label>
        <input :value="outboundAvailableStock" disabled />
      </div>
      <div class="form-group"><label>Дата отгрузки</label>
        <input type="date" v-model="newOutbound.shipped_at" />
      </div>
      <div class="form-group"><label>Расходная накладная / основание</label>
        <input v-model="newOutbound.document_no" placeholder="Например: РН-00045" />
      </div>
      <div class="form-group"><label>Кто отпустил</label>
        <select v-model.number="newOutbound.shipped_by">
          <option disabled value="0">Выберите сотрудника</option>
          <option v-for="u in responsibleUsers" :key="u.user_id" :value="u.user_id">
            {{ u.full_name || u.username }} — {{ roleName(u.role) }}
          </option>
        </select>
      </div>
      <div class="form-group"><label>Комментарий</label>
        <input v-model="newOutbound.note" placeholder="Например: срочная выдача в производство" />
      </div>
      <div class="modal-actions">
        <button @click="confirmAddOutbound">💾 Сохранить</button>
        <button @click="closeAddOutboundModal">❌ Отмена</button>
      </div>
    </div>
  </div>

  <!-- Модалка редактирования -->
  <div v-if="showEditOutboundModal" class="modal-overlay" @click.self="closeEditOutboundModal">
    <div class="modal">
      <h3>Редактировать отгрузку</h3>
      <div class="form-group"><label>Товар</label>
        <select v-model.number="outboundToEdit.item_id">
          <option disabled value="0">Выберите товар</option>
          <option v-for="item in items" :key="item.item_id" :value="item.item_id">{{ item.name }} ({{ item.sku }})</option>
        </select>
      </div>
      <div class="form-group"><label>Со склада</label>
        <select v-model.number="outboundToEdit.warehouse_id">
          <option disabled value="0">Выберите склад</option>
          <option v-for="w in warehouses" :key="w.warehouse_id" :value="w.warehouse_id">{{ w.name }}</option>
        </select>
      </div>
      <div class="form-group"><label>Кому / назначение</label>
        <input v-model="outboundToEdit.destination" />
      </div>
      <div class="form-group"><label>Количество</label>
        <input type="number" min="1" v-model.number="outboundToEdit.quantity" />
      </div>
      <div class="form-group">
        <label>Доступно на складе</label>
        <input :value="editOutboundAvailableStock" disabled />
      </div>
      <div class="form-group"><label>Дата отгрузки</label>
        <input type="date" v-model="outboundToEdit.shipped_at" />
      </div>
      <div class="form-group"><label>Расходная накладная / основание</label>
        <input v-model="outboundToEdit.document_no" />
      </div>
      <div class="form-group"><label>Кто отпустил</label>
        <select v-model.number="outboundToEdit.shipped_by">
          <option disabled value="0">Выберите сотрудника</option>
          <option v-for="u in responsibleUsers" :key="u.user_id" :value="u.user_id">
            {{ u.full_name || u.username }} — {{ roleName(u.role) }}
          </option>
        </select>
      </div>
      <div class="form-group"><label>Комментарий</label>
        <input v-model="outboundToEdit.note" />
      </div>
      <div class="modal-actions">
        <button @click="confirmEditOutbound">💾 Сохранить</button>
        <button @click="closeEditOutboundModal">❌ Отмена</button>
      </div>
    </div>
  </div>
</section>
        <!-- Другое (заглушка) -->
        <section v-if="!tabs.includes(currentTab)">
    <p>Раздел "{{ currentTab }}" в разработке или доступ ограничен...</p>
  </section>
      </main>
    </div>
   <!-- МОДАЛКА ПРОФИЛЯ - вставь в свой <template> -->
    <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
  <div class="modal profile-modal-modern">
    <!-- Заголовок и аватар -->
    <div class="profile-header-modern">
      <div class="profile-avatar-modern accent-avatar">
        <svg width="58" height="58" viewBox="0 0 58 58">
          <defs>
            <linearGradient id="avatar-gradient" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stop-color="#a5b6fa"/>
              <stop offset="100%" stop-color="#2563eb"/>
            </linearGradient>
          </defs>
          <circle cx="29" cy="29" r="29" fill="url(#avatar-gradient)"/>
          <path d="M29 36c-6 0-11 5-11 11h22c0-6-5-11-11-11Zm0-4a7 7 0 1 0 0-14 7 7 0 0 0 0 14Z" fill="#f3f6fd"/>
        </svg>
      </div>
      <div>
        <div class="profile-title-modern main-name">{{ user.full_name }}</div>
        <div class="profile-role-modern">{{ roleName(user.role) }}</div>
      </div>
    </div>

    <!-- Блок с инфой -->
    <div class="profile-info-modern-rich profile-info-compact">
      <div class="info-row-rich">
        <span class="info-icon-circle">
          <svg width="18" height="18" fill="none" viewBox="0 0 18 18"><circle cx="9" cy="9" r="9" fill="#e3eaff"/><path d="M9 12c-2.2 0-4 1.1-4 2.1v.4h8v-.4C13 13.1 11.2 12 9 12zm0-1.1A2.1 2.1 0 1 0 9 6a2.1 2.1 0 0 0 0 4.2z" fill="#2563eb"/></svg>
        </span>
        <span class="info-label-rich">Логин</span>
        <span class="info-value-rich">{{ user.username }}</span>
      </div>
      <div class="info-row-rich">
        <span class="info-icon-circle">
          <svg width="18" height="18" fill="none" viewBox="0 0 18 18"><rect width="18" height="18" rx="6" fill="#e3eaff"/><path d="M5.7 7.2h6.6v1.2H5.7v-1.2zm0 2h6.6v1.2H5.7v-1.2z" fill="#2563eb"/></svg>
        </span>
        <span class="info-label-rich">Роль</span>
        <span class="info-value-rich">{{ roleName(user.role) }}</span>
      </div>
    </div>

    <!-- Блок смены пароля -->
    <div class="profile-info-modern-rich profile-password-compact">
      <div class="profile-change-title-modern">Смена пароля</div>
      <div class="profile-change-fields-modern profile-fields-spaced">
        <input
          type="password"
          v-model="oldPassword"
          placeholder="Старый пароль"
          class="input-modern input-shadow"
        />
        <input
          type="password"
          v-model="newPassword"
          placeholder="Новый пароль"
          class="input-modern input-shadow"
        />
        <input
          type="password"
          v-model="repeatPassword"
          placeholder="Повторите новый пароль"
          class="input-modern input-shadow"
        />
      </div>
      <transition name="fade">
        <div v-if="profileError" class="error-msg-modern">{{ profileError }}</div>
      </transition>
      <transition name="fade">
        <div v-if="profileSuccess" class="success-msg-modern">{{ profileSuccess }}</div>
      </transition>
    </div>

    <!-- Кнопки в самом низу -->
    <div class="profile-actions-modern buttons-bottom profile-actions-outside">
      <button @click="changePassword" class="change-btn-modern main-btn-strong">Сменить пароль</button>
      <button @click="showProfileModal = false" class="close-btn-modern main-btn-ghost">Закрыть</button>
    </div>
  </div>
</div>


  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import BarChart from './components/BarChart.vue'
import LineChart from './components/LineChart.vue'
import { GetWeeklyStockTrend } from '../wailsjs/go/app/App'
import { ChangeStock } from '../wailsjs/go/app/App'
import { GetStockDetails } from '../wailsjs/go/app/App'
import { RemoveStock } from '../wailsjs/go/app/App'
import { ExportStockToExcel } from '../wailsjs/go/app/App'
import { ExportUsersToExcel, ExportSuppliersToExcel, ExportDeliveriesToExcel, ExportItemsToExcel } from '../wailsjs/go/app/App'
import LoginForm from './components/LoginForm.vue'
import {
  GetOutboundDetails, AddOutbound, EditOutbound, RemoveOutbound
} from '../wailsjs/go/app/App'
import {
  GetDashboard,
  GetTopItems,
  GetTurnoverByWarehouse,
  FindStockByWarehouse,
  GetWarehouses,
  AddStock,
  GetAllItems,
  GetInboundDetails,
  GetInboundDetailsByDate,
  AddInbound,
  AddInboundTx,
  GetSuppliers,
  DeleteInbound,
  EditInbound,
  ChangePassword,
  GetItems,
  AddItem,
  UpdateItem,
  RemoveItem,
  EditSupplier,
  AddSupplier,
  RemoveSupplier,
  GetUsers,
  RegisterUser,
  RemoveUser,
  ChangeUserData,
  AddWarehouse,EditWarehouse,
  GetAllMovementsThisMonth,
  TransferStock,
  GetStockTransfers

} from '../wailsjs/go/app/App'

function loadSavedUser() {
  try {
    const saved = localStorage.getItem('currentUser')
    return saved ? JSON.parse(saved) : null
  } catch {
    localStorage.removeItem('currentUser')
    return null
  }
}

const initialUser = loadSavedUser()
const loggedIn = ref(localStorage.getItem('loggedIn') === 'true' && !!initialUser)
const emit = defineEmits(['login-success'])

async function onLoginSuccess(userData) {
  user.value = userData
  loggedIn.value = true
  localStorage.setItem('loggedIn', 'true')
  localStorage.setItem('currentUser', JSON.stringify(userData))
  await loadInitialData()
}

async function loadInitialData() {
  items.value = await GetItems() || []
  weeklyStockData.value = await GetWeeklyStockTrend() || []
  const dash = await GetDashboard()
  totalStock.value = dash.total_stock
  itemCount.value = dash.item_count
  monthlyOrders.value = dash.monthly_orders
  newItems.value = dash.new_items
  suppliers.value = await GetSuppliers() || []
  await loadUsers()
  const stockData = normalizeStockRows(await GetStockDetails())
  stockList.value = stockData
  allStockList.value = stockData
  await reloadMovements()
  await reloadTransfers()
  await reloadOutbound()
  try {
    deliveriesList.value = await GetInboundDetails() || []
  } catch (err) {
    console.error('Ошибка загрузки поставок:', err)
  }
  topItems.value = await GetTopItems() || []
  await loadWarehouses()
  turnoverData.value = await GetTurnoverByWarehouse() || []
}
function logout() {
  localStorage.removeItem('loggedIn')
  localStorage.removeItem('currentUser')
  loggedIn.value = false
  user.value = null
}

// Пользователи
const users = ref([])
const userSearch = ref('')
const userRoleFilter = ref('')
const showAddUserModal = ref(false)
const showEditUserModal = ref(false)
const newUser = ref({ username: '', full_name: '', password: '', role: '' })
const userToEdit = ref({})

const showAddItemModal = ref(false)
const showEditItemModal = ref(false)
const itemToEdit = ref(null)
// Состояния
const outboundList = ref([])
const showAddOutboundModal = ref(false)
const showEditOutboundModal = ref(false)
const outboundToEdit = ref({})
const outboundSearch = ref('')
const outboundDateFilter = ref('')
const outboundWarehouseFilter = ref(0)
const newOutbound = ref({
  item_id: 0,
  warehouse_id: 0,
  destination: '',
  quantity: 1,
  shipped_at: '',
  shipped_by: 0,
  document_no: '',
  note: ''
})

const responsibleUsers = computed(() => {
  if (users.value.length) return users.value
  return user.value ? [user.value] : []
})

function defaultResponsibleId() {
  return user.value?.user_id || responsibleUsers.value[0]?.user_id || 0
}

async function loadUsers() {
  try {
    users.value = await GetUsers() || []
  } catch {
    users.value = user.value ? [user.value] : []
  }
}

const filteredUsers = computed(() =>
  users.value.filter(u => {
    const search = userSearch.value.toLowerCase()
    return (!userRoleFilter.value || u.role === userRoleFilter.value) &&
      (
        (u.username || '').toLowerCase().includes(search) ||
        (u.full_name || '').toLowerCase().includes(search) ||
        (roleName(u.role) || '').toLowerCase().includes(search)
      )
  })
)

const userRoleCounts = computed(() => ({
  admin: users.value.filter(u => u.role === 'admin').length,
  manager: users.value.filter(u => u.role === 'manager').length,
  worker: users.value.filter(u => u.role === 'worker').length
}))

function openAddUserModal() { showAddUserModal.value = true }
function closeAddUserModal() {
  showAddUserModal.value = false
  newUser.value = { username: '', full_name: '', password: '', role: '' }
}

async function confirmAddUser() {
  if (!newUser.value.username || !newUser.value.full_name || !newUser.value.password || !newUser.value.role) {
    alert('Заполните все поля')
    return
  }
  try {
    await RegisterUser(
  newUser.value.username,   // первый аргумент: username
  newUser.value.password,   // второй аргумент: password
  newUser.value.full_name,  // третий аргумент: fullName
  newUser.value.role        // четвертый аргумент: role
)
    users.value = await GetUsers() || []
    closeAddUserModal()
  } catch (e) {
    alert('Ошибка при добавлении пользователя: ' + (e?.message || ''))
  }
}


function openEditUserModal(u) {
  userToEdit.value = { ...u, newPassword: '' }
  showEditUserModal.value = true
}

function closeEditUserModal() { showEditUserModal.value = false }
async function confirmEditUser() {
  if (!userToEdit.value.username || !userToEdit.value.full_name || !userToEdit.value.role) {
    alert('Заполните все поля')
    return
  }
  try {
    const payload = {
      user_id: userToEdit.value.user_id,
      username: userToEdit.value.username,
      full_name: userToEdit.value.full_name,
      role: userToEdit.value.role,
      // Передавай новый пароль только если он не пустой
      password: userToEdit.value.newPassword || undefined
    }
    await ChangeUserData(payload)
    users.value = await GetUsers() || []
    closeEditUserModal()
  } catch (e) {
    alert('Ошибка при обновлении пользователя: ' + (e?.message || ''))
  }
}
const movements = ref([])
const moveType = ref("")
const moveWarehouseId = ref(0)
const moveItemSearch = ref("")
const movementDateFrom = ref("")
const movementDateTo = ref("")
const transferList = ref([])
const showTransferModal = ref(false)
const newTransfer = ref({
  item_id: 0,
  from_warehouse_id: 0,
  to_warehouse_id: 0,
  quantity: 1,
  transferred_at: '',
  note: ''
})

// График - динамика по дням (поступления и отгрузки)
const moveChartData = computed(() => {
  // Группируем движения по дате, считаем поступления/отгрузки отдельно
  const byDate = {}
  for (const m of filteredMovements.value) {
    const date = (new Date(m.date)).toISOString().slice(0, 10)
    if (!byDate[date]) byDate[date] = { in: 0, out: 0 }
    if (m.type === 'inbound') byDate[date].in += m.quantity
    if (m.type === 'outbound') byDate[date].out += m.quantity
  }
  const dates = Object.keys(byDate).sort()
  return {
    labels: dates.map(d => d.split('-').reverse().join('.')),
    datasets: [
      {
        label: "Поступления",
        data: dates.map(d => byDate[d].in),
        borderColor: "#22c55e",
        backgroundColor: "rgba(34,197,94,0.1)",
        tension: 0.3
      },
      {
        label: "Отгрузки",
        data: dates.map(d => byDate[d].out),
        borderColor: "#ef4444",
        backgroundColor: "rgba(239,68,68,0.1)",
        tension: 0.3
      }
    ]
  }
})

const filteredMovements = computed(() =>
  movements.value.filter(m => {
    const wid = Number(moveWarehouseId.value)  // ← здесь
    const date = normalizeDateInput(m.date)
    const fromOk = !movementDateFrom.value || date >= movementDateFrom.value
    const toOk = !movementDateTo.value || date <= movementDateTo.value
    const search = moveItemSearch.value.toLowerCase()
    return (!moveType.value || m.type === moveType.value)
        && (wid === 0 || m.warehouse_id === wid)
        && fromOk
        && toOk
        && (
            (m.item_name || '').toLowerCase().includes(search) ||
            (m.warehouse_name || '').toLowerCase().includes(search) ||
            (m.supplier_name || '').toLowerCase().includes(search) ||
            (m.destination || '').toLowerCase().includes(search) ||
            m.item_id.toString().includes(search)
        )
  })
)


const inboundCount  = computed(() =>
  filteredMovements.value.filter(m => m.type === 'inbound').length
)

const outboundCount = computed(() =>
  filteredMovements.value.filter(m => m.type === 'outbound').length)
const transferCount = computed(() =>
  filteredMovements.value.filter(m => m.type === 'transfer').length
)
const totalMoved = computed(() => filteredMovements.value.reduce((acc, m) => acc + m.quantity, 0))

const filteredTransferList = computed(() => {
  const search = moveItemSearch.value.toLowerCase()
  return transferList.value.filter(t => {
    const date = normalizeDateInput(t.date)
    return (!movementDateFrom.value || date >= movementDateFrom.value) &&
      (!movementDateTo.value || date <= movementDateTo.value) &&
      (!search ||
        (t.item_name || '').toLowerCase().includes(search) ||
        (t.sku || '').toLowerCase().includes(search) ||
        (t.from_warehouse || '').toLowerCase().includes(search) ||
        (t.to_warehouse || '').toLowerCase().includes(search) ||
        (t.note || '').toLowerCase().includes(search)
      )
  })
})

function movementTypeName(t) {
  if (t === 'inbound') return 'Поступление'
  if (t === 'outbound') return 'Отгрузка'
  if (t === 'transfer') return 'Перемещение'
  return t
}


async function reloadMovements() {
  // Можно сюда добавить аргументы фильтров, если делаешь серверный фильтр
  movements.value = await GetAllMovementsThisMonth() || []
}

async function reloadTransfers() {
  transferList.value = await GetStockTransfers() || []
}

function openTransferModal() {
  newTransfer.value = {
    item_id: 0,
    from_warehouse_id: 0,
    to_warehouse_id: 0,
    quantity: 1,
    transferred_at: '',
    note: ''
  }
  showTransferModal.value = true
}

function openTransferFromStock(stock) {
  newTransfer.value = {
    item_id: stock.item_id,
    from_warehouse_id: stock.warehouse_id,
    to_warehouse_id: 0,
    quantity: 1,
    transferred_at: '',
    note: `Перемещение ${stock.sku}`
  }
  currentTab.value = 'Движения'
  showTransferModal.value = true
}

function closeTransferModal() {
  showTransferModal.value = false
}

const transferAvailableStock = computed(() => {
  if (!newTransfer.value.item_id || !newTransfer.value.from_warehouse_id) return 0
  return allStockList.value
    .filter(row =>
      row.item_id === newTransfer.value.item_id &&
      row.warehouse_id === newTransfer.value.from_warehouse_id
    )
    .reduce((acc, row) => acc + Number(row.quantity || 0), 0)
})

async function confirmTransfer() {
  const payload = newTransfer.value
  if (!payload.item_id || !payload.from_warehouse_id || !payload.to_warehouse_id || !payload.quantity) {
    alert('Заполните товар, склады и количество')
    return
  }
  if (payload.from_warehouse_id === payload.to_warehouse_id) {
    alert('Склады отправления и получения должны отличаться')
    return
  }
  if (payload.quantity <= 0) {
    alert('Количество должно быть больше 0')
    return
  }
  if (payload.quantity > transferAvailableStock.value) {
    alert(`Недостаточно остатка на складе-источнике. Доступно: ${transferAvailableStock.value}`)
    return
  }

  try {
    await TransferStock(
      payload.item_id,
      payload.from_warehouse_id,
      payload.to_warehouse_id,
      payload.quantity,
      payload.transferred_at,
      payload.note
    )
    closeTransferModal()
    await refreshAllStockDetails()
    await reloadTransfers()
    await reloadMovements()
    await reloadWeeklyTrend()
  } catch (e) {
    alert('Ошибка при перемещении: ' + (e?.message || e))
  }
}

async function deleteUser(u) {
  if (!confirm(`Удалить пользователя "${u.username}"?`)) return
  try {
    await RemoveUser(u.user_id)
    users.value = await GetUsers() || []
  } catch (e) {
    alert('Ошибка при удалении: ' + (e?.message || ''))
  }
}

function exportUsersToExcel() {
  ExportUsersToExcel().then(base64data => {
    const binary = atob(base64data)
    const len = binary.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = 'users.xlsx'
    link.click()
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
  }).catch(err => {
    alert('Ошибка экспорта: ' + err)
  })
}

function handleLogin() {
  error.value = ''
  loading.value = true
  setTimeout(() => {
    if (login.value === 'admin' && password.value === '1234') {
      localStorage.setItem('loggedIn', 'true')
      emit('login-success')
    } else {
      error.value = 'Неверный логин или пароль'
    }
    loading.value = false
  }, 700)
}

const user = ref(initialUser) // по умолчанию null
// Подгружай реального пользователя с бэка после логина
function roleName(role) {
  switch (role) {
    case 'admin': return 'Администратор'
    case 'manager': return 'Менеджер'
    case 'worker': return 'Сотрудник'
    default: return 'Пользователь'
  }
}
const showProfileModal = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const repeatPassword = ref('')
const profileError = ref('')
const profileSuccess = ref('')

const tabs = computed(() => {
  if (!user.value) return [];
  if (user.value.role === 'admin') {
    return [
      'Дашборд',
      'Остатки',
      'Поставки',
      'Отгрузки',
      'План закупок',
      'Товары',
      'Склады',
      'Поставщики',
      'Движения',
      'Пользователи' // Только для админа!
    ];
  }
  // Для менеджера
  if (user.value.role === 'manager') {
    return [
      'Дашборд',
      'Остатки',
      'Поставки',
      'Отгрузки',
      'План закупок',
      'Товары',
      'Склады',
      'Поставщики',
      'Движения'
    ];
  }
  // Для сотрудника
  return [
    'Дашборд',
    'Остатки',
    'Поставки',
    'Отгрузки',
    'Движения'
  ];
});

async function changePassword() {
  profileError.value = ''
  profileSuccess.value = ''
  if (!oldPassword.value || !newPassword.value || !repeatPassword.value) {
    profileError.value = 'Заполните все поля'
    return
  }
  if (newPassword.value !== repeatPassword.value) {
    profileError.value = 'Пароли не совпадают'
    return
  }
  try {
    // используем username (логин)
    await window.go.app.App.ChangePassword(user.value.username, oldPassword.value, newPassword.value)
    profileSuccess.value = 'Пароль успешно изменён'
    oldPassword.value = newPassword.value = repeatPassword.value = ''
  } catch (e) {
    profileError.value = e?.message || 'Ошибка смены пароля'
  }
}

const showEditDeliveryModal = ref(false)
const deliveryToEdit = ref(null)
const showEditModal = ref(false)
const stockToEdit = ref(null)
const warehouses = ref([])
const selectedWarehouseId = ref(0)
const searchQuery = ref('')
const stockStatusFilter = ref('all')
const currentTab = ref('Дашборд')
const stockList = ref([])
const allStockList = ref([])
const totalStock = ref(0)
const itemCount = ref(0)
const monthlyOrders = ref(0)
const newItems = ref(0)
const topItems = ref([])
const turnoverData = ref([])
const weeklyStockData = ref([])
const showAddDeliveryModal = ref(false)
const newInbound = ref({
  item_id: 0,
  supplier_id: 0,
  warehouse_id: 0,
  quantity: 1,
  received_at: "",
  received_by: 0,
  document_no: '',
  note: '',
})
const newInboundItem = ref({
  sku: '',
  name: '',
  description: '',
  uom: '',
  price: 0,
  cost: 0,
})
const suppliers = ref([])
const selectedRiskWarehouseId = ref(0)
const purchasePlanSearch = ref('')
const purchasePlanSupplierId = ref(0)
const purchasePlanPriority = ref('all')
const purchasePlanHorizon = ref(30)
const selectedPurchaseKeys = ref(new Set())
 // если вкладки реализованы через состояние, замени на свой способ
const warehouseSearch = ref('')
const newWarehouse = ref({ name: '', location: '' })
const editWarehouseData = ref({ warehouse_id: null, name: '', location: '' })
const supplierSearch = ref('');
const supplierActivityFilter = ref('all');
const filteredSuppliers = computed(() =>
  suppliers.value.filter(s => {
    const stats = supplierStats(s.supplier_id)
    const search = supplierSearch.value.toLowerCase()
    const matchesActivity =
      supplierActivityFilter.value === 'all' ||
      (supplierActivityFilter.value === 'active' && stats.deliveries > 0) ||
      (supplierActivityFilter.value === 'idle' && stats.deliveries === 0)

    return matchesActivity &&
      (
        (s.name || '').toLowerCase().includes(search) ||
        (s.inn || '').toLowerCase().includes(search) ||
        (s.contact_person || '').toLowerCase().includes(search)
      )
  })
);
const showAddSupplierModal = ref(false);
const showEditSupplierModal = ref(false);
const newSupplier = ref({ name: '', inn: '', contact_person: '', phone: '', email: '' });
const supplierToEdit = ref({});

async function loadWarehouses() {
  warehouses.value = await GetWarehouses() || []
}


function openAddSupplierModal() { showAddSupplierModal.value = true }
async function confirmAddSupplier() {
  // Валидация при необходимости
  try {
    await window.go.app.App.AddSupplier(newSupplier.value)
    showAddSupplierModal.value = false
    newSupplier.value = { name: '', inn: '', contact_person: '', phone: '', email: '' }
    // После добавления — обновить список с бэка
    suppliers.value = await GetSuppliers() || []
  } catch (e) {
    alert('Ошибка при добавлении: ' + (e?.message || ''))
  }
}

async function confirmEditSupplier() {
  try {
    await window.go.app.App.EditSupplier(supplierToEdit.value)
    showEditSupplierModal.value = false
    suppliers.value = await GetSuppliers() || []
  } catch (e) {
    alert('Ошибка при обновлении: ' + (e?.message || ''))
  }
}

function openEditSupplierModal(s) { supplierToEdit.value = { ...s }; showEditSupplierModal.value = true }

async function deleteSupplier(s) {
  if (!confirm(`Удалить поставщика "${s.name}"?`)) return
  try {
    // Удаляем по уникальному supplier_id
    await window.go.app.App.RemoveSupplier(s.supplier_id)
    // После удаления — обновить список с бэка
    suppliers.value = await GetSuppliers() || []
  } catch (e) {
    alert('Ошибка при удалении: ' + (e?.message || ''))
  }
}

function maskPhone(event, obj) {
  let v = event.target.value.replace(/\D/g, '');
  if (v.startsWith('8')) v = '7' + v.slice(1); // Преобразуем 8 -> 7
  if (!v.startsWith('7')) v = '7' + v;
  v = v.slice(0, 11);

  let res = '+7';
  if (v.length > 1) res += ' (' + v.slice(1, 4);
  if (v.length >= 4) res += ')';
  if (v.length >= 4) res += '-' + v.slice(4, 7);
  if (v.length >= 7) res += '-' + v.slice(7, 9);
  if (v.length >= 9) res += '-' + v.slice(9, 11);
  obj.phone = res;
}


function exportSuppliersToExcel() {
  ExportSuppliersToExcel().then(base64data => {
    const binary = atob(base64data)
    const len = binary.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = 'suppliers.xlsx'
    link.click()
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
  }).catch(err => {
    alert('Ошибка экспорта: ' + err)
  })
}

function exportToExcel() {
  window.go.app.App.ExportStockToExcel().then(base64data => {
    const binary = atob(base64data);
    const len = binary.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i);
    }

    const blob = new Blob([bytes], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = "stock_report.xlsx";
    link.click();
    setTimeout(() => URL.revokeObjectURL(link.href), 1000);
  }).catch(err => {
    alert("Ошибка экспорта: " + err);
  });
}

function openAddDeliveryModal() {
  showAddDeliveryModal.value = true
  newInbound.value = {
    item_id: 0,
    supplier_id: 0,
    warehouse_id: 0,
    quantity: 1,
    received_at: '',
    received_by: defaultResponsibleId(),
    document_no: '',
    note: ''
  }
  Object.assign(newInboundItem.value, { sku: '', name: '', description: '', uom: '', price: 0, cost: 0 })
}

function startInboundFromPlan(item) {
  const warehouseId = item.warehouse_id || Number(selectedRiskWarehouseId.value) || warehouses.value[0]?.warehouse_id || 0
  currentTab.value = 'Поставки'
  newInbound.value = {
    item_id: item.item_id,
    supplier_id: item.supplier_id || purchasePlanSupplierId.value || 0,
    warehouse_id: warehouseId,
    quantity: item.recommended_order || 1,
    received_at: '',
    received_by: defaultResponsibleId(),
    document_no: '',
    note: 'Создано из плана закупок'
  }
  Object.assign(newInboundItem.value, { sku: '', name: '', description: '', uom: '', price: 0, cost: 0 })
  showAddDeliveryModal.value = true
}

function normalizeStockRows(rows) {
  return (rows || []).map(s => ({
    id: s.stock_id,
    stock_id: s.stock_id,
    item_id: s.item_id,
    warehouse_id: s.warehouse_id,
    name: s.name,
    sku: s.sku,
    warehouse: s.warehouse,
    quantity: s.quantity
  }))
}

function numericValue(value) {
  if (value === null || value === undefined || value === '') return 0
  if (typeof value === 'number') return value
  if (typeof value === 'object') {
    if (value.Valid === false) return 0
    if (value.Float64 !== undefined) return Number(value.Float64) || 0
  }
  return Number(value) || 0
}

function itemPrice(item) {
  return numericValue(item?.price)
}

function itemCost(item) {
  return numericValue(item?.cost)
}

function formatMoneyValue(value) {
  const number = Number(value) || 0
  return number > 0 ? number.toLocaleString('ru-RU', { minimumFractionDigits: 2 }) : '—'
}

function formatNumber(value, digits = 0) {
  return new Intl.NumberFormat('ru-RU', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits
  }).format(Number(value) || 0)
}

function normalizeDateInput(value) {
  if (!value) return ''
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}/.test(value)) return value.slice(0, 10)
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toISOString().slice(0, 10)
}

function itemStockTotal(itemID) {
  return allStockList.value
    .filter(row => row.item_id === itemID)
    .reduce((acc, row) => acc + Number(row.quantity || 0), 0)
}

function stockAvailable(itemID, warehouseID) {
  if (!itemID || !warehouseID) return 0
  return allStockList.value
    .filter(row => row.item_id === Number(itemID) && row.warehouse_id === Number(warehouseID))
    .reduce((acc, row) => acc + Number(row.quantity || 0), 0)
}

function itemMarginPercent(item) {
  const price = itemPrice(item)
  const cost = itemCost(item)
  if (!price || !cost) return 0
  return Math.round(((price - cost) / price) * 100)
}

function csvEscape(value) {
  const raw = String(value ?? '')
  return `"${raw.replace(/"/g, '""')}"`
}

function downloadCsv(filename, headers, rows) {
  const csv = [
    headers.map(csvEscape).join(';'),
    ...rows.map(row => row.map(csvEscape).join(';'))
  ].join('\n')
  const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
  setTimeout(() => URL.revokeObjectURL(link.href), 1000)
}

async function refreshAllStockDetails() {
  const rows = normalizeStockRows(await GetStockDetails())
  allStockList.value = rows
  const warehouseId = Number(selectedWarehouseId.value)
  stockList.value = warehouseId === 0
    ? rows
    : rows.filter(row => row.warehouse_id === warehouseId)
}

const averagePrice = computed(() => {
  if (!items.value.length) return '—'
  // Игнорируем товары без цены (null или 0 можно убрать по желанию)
  const filtered = items.value.filter(i => itemPrice(i) > 0)
  if (!filtered.length) return '—'
  const sum = filtered.reduce((acc, i) => acc + itemPrice(i), 0)
  return (sum / filtered.length).toLocaleString('ru-RU', { minimumFractionDigits: 2 })
})

const averageMarginPercent = computed(() => {
  const margins = items.value
    .map(itemMarginPercent)
    .filter(value => value > 0)
  if (!margins.length) return 0
  return Math.round(margins.reduce((acc, value) => acc + value, 0) / margins.length)
})

  function openEditDeliveryModal(delivery) {
  let date = delivery.received_at
  if (typeof date === "string" && date.includes(".")) {
    // DD.MM.YYYY -> YYYY-MM-DD
    const [dd, mm, yyyy] = date.split(".");
    date = `${yyyy}-${mm}-${dd}`;
  }
  deliveryToEdit.value = {
    ...delivery,
    received_at: date ? date.substring(0, 10) : "",
    received_by: delivery.received_by || defaultResponsibleId(),
    document_no: delivery.document_no || '',
    note: delivery.note || ''
  }
  showEditDeliveryModal.value = true
}

// Закрыть модалку
function closeEditDeliveryModal() {
  showEditDeliveryModal.value = false
  deliveryToEdit.value = null
}

// Подтвердить редактирование
function confirmEditDelivery() {
  if (
    !deliveryToEdit.value.item_id ||
    !deliveryToEdit.value.supplier_id ||
    !deliveryToEdit.value.warehouse_id ||
    !deliveryToEdit.value.quantity ||
    deliveryToEdit.value.quantity <= 0 ||
    !deliveryToEdit.value.received_at ||
    !deliveryToEdit.value.received_by
  ) {
    alert("Заполните все поля");
    return;
  }
  // Собираем payload для Go backend
  const receivedAt = deliveryToEdit.value.received_at
  ? new Date(deliveryToEdit.value.received_at).toISOString()
  : undefined;

const payload = {
  inbound_id: deliveryToEdit.value.inbound_id,
  item_id: deliveryToEdit.value.item_id,
  supplier_id: deliveryToEdit.value.supplier_id,
  warehouse_id: deliveryToEdit.value.warehouse_id,
  quantity: deliveryToEdit.value.quantity,
  received_at: receivedAt,
  received_by: deliveryToEdit.value.received_by || defaultResponsibleId(),
  document_no: deliveryToEdit.value.document_no || null,
  note: deliveryToEdit.value.note || null
}
    window.go.app.App.EditInbound(payload).then(() => {
    closeEditDeliveryModal()
    GetInboundDetails().then(data => {
      deliveriesList.value = data || []
    })
    refreshAllStockDetails()
  }).catch(err => {
    alert("Ошибка при обновлении поставки")
    console.error(err)
  })
}


function closeAddDeliveryModal() {
  showAddDeliveryModal.value = false
  newInbound.value = {
    item_id: 0,
    supplier_id: 0,
    warehouse_id: 0,
    quantity: 1,
    received_at: "",
    received_by: defaultResponsibleId(),
    document_no: '',
    note: '',
  }
  Object.assign(newInboundItem.value, { sku: '', name: '', description: '', uom: '', price: 0, cost: 0 })
}
const outboundChartData = computed(() => {
  const byDate = {}
  for (const o of outboundList.value) {
    const date = (new Date(o.date)).toISOString().slice(0, 10)
    byDate[date] = (byDate[date] || 0) + o.quantity
  }
  const dates = Object.keys(byDate).sort()
  return {
    labels: dates.map(d => d.split('-').reverse().join('.')),
    datasets: [{
      label: "Отгрузки",
      data: dates.map(d => byDate[d]),
      borderColor: "#ef4444",
      backgroundColor: "rgba(239,68,68,0.15)",
      tension: 0.3
    }]
  }
})

const filteredOutboundList = computed(() =>
  (outboundList.value || []).filter(o =>
    (!outboundDateFilter.value || (o.date || '').startsWith(outboundDateFilter.value)) &&
    (Number(outboundWarehouseFilter.value) === 0 || o.warehouse_id === Number(outboundWarehouseFilter.value)) &&
    (
      (o.name || '').toLowerCase().includes(outboundSearch.value.toLowerCase()) ||
      (o.sku || '').toLowerCase().includes(outboundSearch.value.toLowerCase()) ||
      (o.warehouse || '').toLowerCase().includes(outboundSearch.value.toLowerCase()) ||
      (o.destination || '').toLowerCase().includes(outboundSearch.value.toLowerCase()) ||
      (o.document_no || '').toLowerCase().includes(outboundSearch.value.toLowerCase()) ||
      (o.shipper_name || '').toLowerCase().includes(outboundSearch.value.toLowerCase())
    )
  )
)


const totalOutboundQuantity = computed(() => filteredOutboundList.value.reduce((acc, o) => acc + Number(o.quantity || 0), 0))
const outboundDestinationCount = computed(() =>
  new Set(filteredOutboundList.value.map(o => o.destination).filter(Boolean)).size
)
const averageOutboundSize = computed(() =>
  filteredOutboundList.value.length ? totalOutboundQuantity.value / filteredOutboundList.value.length : 0
)
const outboundAvailableStock = computed(() => stockAvailable(newOutbound.value.item_id, newOutbound.value.warehouse_id))
const editOutboundAvailableStock = computed(() => stockAvailable(outboundToEdit.value.item_id, outboundToEdit.value.warehouse_id) + Number(outboundToEdit.value.quantity || 0))
// Эту функцию вызывай при сохранении поставки
function confirmAddDelivery() {
  // Простая валидация
  if (
    !newInbound.value.item_id ||
    !newInbound.value.supplier_id ||
    !newInbound.value.warehouse_id ||
    !newInbound.value.quantity ||
    newInbound.value.quantity <= 0 ||
    !newInbound.value.received_by
  ) {
    alert("Заполните все обязательные поля");
    return;
  }
  const receivedAt = newInbound.value.received_at
  ? new Date(newInbound.value.received_at).toISOString()
  : undefined;
  // Если выбран новый товар
  if (newInbound.value.item_id === -1) {
    if (!newInboundItem.value.sku || !newInboundItem.value.name) {
      alert('Введите SKU и наименование товара');
      return;
    }
    const inboundPayload = {
      item_id: 0,
      supplier_id: newInbound.value.supplier_id,
      warehouse_id: newInbound.value.warehouse_id,
      quantity: newInbound.value.quantity,
      received_at: receivedAt,
      received_by: newInbound.value.received_by || defaultResponsibleId(),
      document_no: newInbound.value.document_no || null,
      note: newInbound.value.note || null,
    }
    const itemPayload = {
      sku: newInboundItem.value.sku,
      name: newInboundItem.value.name,
      description: newInboundItem.value.description,
      uom: newInboundItem.value.uom || 'шт',
      reorder_level: 0,
      reorder_qty: 0,
      price: newInboundItem.value.price,
      cost: newInboundItem.value.cost,
    }
    window.go.app.App.AddInboundTx(inboundPayload, itemPayload).then(() => {
      closeAddDeliveryModal()
      GetInboundDetails().then(data => { deliveriesList.value = data || [] })
      GetItems().then(data => { items.value = data || [] })
      refreshAllStockDetails()
    }).catch(err => {
      alert('Ошибка при добавлении поставки')
      console.error(err)
    })
  } else {
    // Если дата не выбрана, на бэке ставится now()
    const payload = {
      item_id: newInbound.value.item_id,
      supplier_id: newInbound.value.supplier_id,
      warehouse_id: newInbound.value.warehouse_id,
      quantity: newInbound.value.quantity,
      received_at: receivedAt,
      received_by: newInbound.value.received_by || defaultResponsibleId(),
      document_no: newInbound.value.document_no || null,
      note: newInbound.value.note || null
    }
    window.go.app.App.AddInbound(payload).then(() => {
      closeAddDeliveryModal()
      // обновить deliveriesList после добавления
      GetInboundDetails().then(data => {
        deliveriesList.value = data || []
      })
      refreshAllStockDetails()
    }).catch(err => {
      alert("Ошибка при добавлении поставки")
      console.error(err)
    })
  }
}
async function reloadOutbound() {
  const data = await GetOutboundDetails();
  outboundList.value = Array.isArray(data) ? data : [];
}


watch(outboundDateFilter, reloadOutbound)

// Добавление
function openAddOutboundModal() {
  showAddOutboundModal.value = true
  newOutbound.value = {
    item_id: 0,
    warehouse_id: 0,
    destination: '',
    quantity: 1,
    shipped_at: '',
    shipped_by: defaultResponsibleId(),
    document_no: '',
    note: ''
  }
}
function closeAddOutboundModal() { showAddOutboundModal.value = false }
async function confirmAddOutbound() {
  if (!newOutbound.value.item_id || !newOutbound.value.warehouse_id || !newOutbound.value.destination || !newOutbound.value.quantity || !newOutbound.value.shipped_by) {
    alert('Заполните все поля')
    return
  }
  if (newOutbound.value.quantity > outboundAvailableStock.value) {
    alert(`Недостаточно остатка на складе. Доступно: ${outboundAvailableStock.value}`)
    return
  }
  try {
    await AddOutbound(
  newOutbound.value.item_id,
  newOutbound.value.quantity,
  newOutbound.value.shipped_at,   // строка YYYY-MM-DD
  newOutbound.value.destination,
  newOutbound.value.warehouse_id,
  newOutbound.value.shipped_by,
  newOutbound.value.document_no,
  newOutbound.value.note
)
    closeAddOutboundModal()
    await  reloadOutbound()
    await reloadMovements()
    await refreshAllStockDetails()
  } catch (e) {
    alert('Ошибка при добавлении: ' + (e?.message || ''))
  }
}

// Редактирование
function openEditOutboundModal(o) {
  outboundToEdit.value = {
    ...o,
    shipped_by: o.shipped_by || defaultResponsibleId(),
    document_no: o.document_no || '',
    note: o.note || ''
  }
  showEditOutboundModal.value = true
}
function closeEditOutboundModal() { showEditOutboundModal.value = false }
async function confirmEditOutbound() {
  // в вашем outboundToEdit.value уже есть все поля:
  const o = outboundToEdit.value;

  // валидация...
  if (o.quantity > editOutboundAvailableStock.value) {
    alert(`Недостаточно остатка на складе. Доступно: ${editOutboundAvailableStock.value}`)
    return
  }
  try {
    // передаём 6 отдельных параметров в том же порядке,
    // в каком вы объявили метод в app.go
    await EditOutbound(
      o.item_id,         // 1) itemID
      o.quantity,        // 2) quantity
      o.shipped_at,      // 3) shippedAtStr (YYYY-MM-DD)
      o.destination,     // 4) destination
      o.warehouse_id,    // 5) warehouseID
      o.outbound_id,     // 6) outboundID
      o.shipped_by || defaultResponsibleId(),
      o.document_no || '',
      o.note || ''
    );

    closeEditOutboundModal();
    await reloadOutbound();
    await reloadMovements();
    await refreshAllStockDetails();
  } catch (e) {
    alert('Ошибка при сохранении: ' + e?.message);
  }
}

// Удаление
async function deleteOutbound(o) {
  if (!confirm(`Удалить отгрузку товара "${o.name}"?`)) return
  try {
    await RemoveOutbound(o.outbound_id)
    reloadOutbound()
    refreshAllStockDetails()
  } catch (e) {
    alert('Ошибка при удалении: ' + (e?.message || ''))
  }
}
const weeklyStockChartData = computed(() => ({
  labels: weeklyStockData.value.map(d => formatDate(d.date)),
  datasets: [
    {
      label: 'Остатки',
      data: weeklyStockData.value.map(d => Number(d.total)),
      backgroundColor: 'rgba(0, 0, 0, 0.2)',   // полупрозрачный чёрный
      borderColor: '#000',                    // чёрная линия
      pointBackgroundColor: '#000',           // чёрные точки
      pointRadius: (ctx) => ctx.dataIndex === weeklyStockData.value.length - 1 ? 6 : 0,
      pointHoverRadius: 6,
      borderWidth: 3,
      fill: true,
      tension: 0.3
    }
  ]
}))

function deleteStock(stock) {
  if (!confirm(`Удалить остаток товара "${stock.name}" со склада "${stock.warehouse}"?`)) {
    return
  }

  RemoveStock(stock.stock_id)
    .then(() => {
      refreshAllStockDetails()
      const reload = selectedWarehouseId.value === 0
        ? GetStockDetails
        : () => FindStockByWarehouse(selectedWarehouseId.value)

      reload().then(data => {
        stockList.value = normalizeStockRows(data)
      })
    })
    .catch(err => {
      alert("Ошибка при удалении")
      console.error(err)
    })
}


const selectedDeliveryWarehouse = ref(0)
const selectedDeliverySupplier = ref(0)
const selectedDeliveryDate = ref("")
const deliverySearchQuery = ref("")
const deliveriesList = ref([])

watch(selectedDeliveryDate, (date) => {
  if (date) {
    GetInboundDetailsByDate(date).then(data => {
      deliveriesList.value = data || [];
    });
  } else {
    GetInboundDetails().then(data => {
      deliveriesList.value = data || [];
    });
  }
});

const filteredDeliveriesList = computed(() =>
  deliveriesList.value.filter(d => {
    const search = deliverySearchQuery.value.toLowerCase()
    return (Number(selectedDeliveryWarehouse.value) === 0 || d.warehouse_id === Number(selectedDeliveryWarehouse.value)) &&
      (Number(selectedDeliverySupplier.value) === 0 || d.supplier_id === Number(selectedDeliverySupplier.value)) &&
      (
        (d.name || '').toLowerCase().includes(search) ||
        (d.sku || '').toLowerCase().includes(search) ||
        (d.supplier || '').toLowerCase().includes(search) ||
        (d.document_no || '').toLowerCase().includes(search) ||
        (d.receiver_name || '').toLowerCase().includes(search) ||
        (d.note || '').toLowerCase().includes(search)
      )
  })
)

const filteredDeliveryUnits = computed(() =>
  filteredDeliveriesList.value.reduce((acc, d) => acc + Number(d.quantity || 0), 0)
)

const filteredDeliverySupplierCount = computed(() =>
  new Set(filteredDeliveriesList.value.map(d => d.supplier_id).filter(Boolean)).size
)

const averageDeliverySize = computed(() =>
  filteredDeliveriesList.value.length ? filteredDeliveryUnits.value / filteredDeliveriesList.value.length : 0
)

const filteredDeliveriesChartData = computed(() => ({
  labels: filteredDeliveriesList.value.map(d => d.name),
  datasets: [
    {
      label: 'Поставки',
      data: filteredDeliveriesList.value.map(d => d.quantity),
      backgroundColor: '#000',
      barThickness: 20
    }
  ]
}))

function deleteDelivery(delivery) {
  console.log("Удаляем поставку:", delivery);
  if (confirm(`Удалить поставку "${delivery.name}" от "${delivery.supplier}"?`)) {
    window.go.app.App.DeleteInbound(delivery.inbound_id || delivery.id)
      .then(() => {
        // обнови deliveriesList
        GetInboundDetails().then(data => {
          deliveriesList.value = data || [];
        });
        refreshAllStockDetails()
      })
      .catch(err => {
        alert("Ошибка при удалении поставки");
        console.error(err);
      });
  }
}

function exportDeliveriesToExcel() {
  ExportDeliveriesToExcel().then(base64data => {
    const binary = atob(base64data)
    const len = binary.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = 'deliveries.xlsx'
    link.click()
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
  }).catch(err => {
    alert('Ошибка экспорта: ' + err)
  })
}

// Форматирование даты, чтобы не падало

const filteredChartData = computed(() => {
  return {
    labels: filteredStockList.value.map(item => item.name),
    datasets: [
      {
        label: 'Остатки',
        data: filteredStockList.value.map(item => item.quantity),
        backgroundColor: '#000',
        barThickness: 20
      }
    ]
  }
})

const turnoverBarChartData = computed(() => ({
  labels: turnoverData.value.map(item => item.name),
  datasets: [
    {
      label: 'Поступило',
      data: turnoverData.value.map(item => item.received),
      backgroundColor: '#000',
      barThickness: 20
    }
  ]
}))

const turnoverLineChartData = computed(() => ({
  labels: turnoverData.value.map(d => d.name),
  datasets: [
    {
      label: 'Оборот',
      data: turnoverData.value.map(d => d.received),
      borderColor: '#000',
      tension: 0.4,
      fill: false
    }
  ]
}))

function formatDate(dateStr) {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  if (isNaN(d)) return dateStr;
  const dd = String(d.getDate()).padStart(2, '0');
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const yyyy = d.getFullYear();
  return `${dd}.${mm}.${yyyy}`;
}

function openEditModal(stock) {
  const warehouse = warehouses.value.find(w => w.name === stock.warehouse)

  stockToEdit.value = {
    ...stock,
    item_id: stock.item_id,
    warehouse_id: warehouse?.warehouse_id
  }

  showEditModal.value = true
}

function confirmEditStock() {
  if (!stockToEdit.value || stockToEdit.value.quantity < 0) {
    alert("Количество должно быть больше 0")
    return
  }

  // item_id и warehouse_id обязательно нужны
  ChangeStock(
  stockToEdit.value.item_id,
  stockToEdit.value.warehouse_id,
  stockToEdit.value.quantity
)
    .then(() => {
      closeEditModal()
      reloadWeeklyTrend()
      refreshAllStockDetails()
      if (selectedWarehouseId.value === 0) {
        GetStockDetails().then(data => {
  stockList.value = normalizeStockRows(data)
})      } else {
        FindStockByWarehouse(selectedWarehouseId.value).then(data => {
          stockList.value = data.map(s => ({
            id: s.item_id,
            name: s.name,
            sku: s.sku,
            warehouse: warehouses.value.find(w => w.warehouse_id === selectedWarehouseId.value)?.name || '',
            quantity: s.quantity
          }))
        })
      }
    })
    .catch(err => {
      alert("Ошибка при обновлении остатка")
      console.error(err)
    })
}

const showAddModal = ref(false)
const newStock = ref({ item_id: 0, warehouse_id: 0, quantity: 0 })
const items = ref([])

function openAddModal() {
  showAddModal.value = true
}
// Категории (генерируются из items)
const categories = computed(() => {
  return Array.from(new Set(items.value.map(i => i.category)));
});

// Поиск и фильтр
const itemSearch = ref('');
const selectedCategory = ref('');
const itemStatusFilter = ref('all');

const filteredItems = computed(() =>
  items.value.filter(i => {
    const search = itemSearch.value.toLowerCase()
    const total = itemStockTotal(i.item_id)
    const reorderLevel = Number(i.reorder_level) || 0
    const price = itemPrice(i)
    const cost = itemCost(i)
    const matchesStatus =
      itemStatusFilter.value === 'all' ||
      (itemStatusFilter.value === 'no_stock' && total === 0) ||
      (itemStatusFilter.value === 'low' && reorderLevel > 0 && total <= reorderLevel) ||
      (itemStatusFilter.value === 'profitable' && price > cost && cost > 0) ||
      (itemStatusFilter.value === 'no_price' && price === 0)

    return matchesStatus &&
      (
        (i.name || '').toLowerCase().includes(search) ||
        (i.sku || '').toLowerCase().includes(search) ||
        (i.description || '').toLowerCase().includes(search)
      )
  })
);


// Минимальный и максимальный остаток
const minStock = computed(() => {
  if (items.value.length === 0) return 0;
  return Math.min(...items.value.map(i => i.min_stock || 0));
});
const maxStock = computed(() => {
  if (items.value.length === 0) return 0;
  return Math.max(...items.value.map(i => i.min_stock || 0));
});

// Заглушки методов
const minActualStock = computed(() => {
  if (!totalStockPerItem.value.length) return 0
  return Math.min(...totalStockPerItem.value)
})
const maxActualStock = computed(() => {
  if (!totalStockPerItem.value.length) return 0
  return Math.max(...totalStockPerItem.value)
})


const totalStockPerItem = computed(() => {
  const map = {}
  for (const s of allStockList.value) {
    map[s.item_id] = (map[s.item_id] || 0) + s.quantity
  }
  return Object.values(map)
})

function exportItemsToExcel() {
  ExportItemsToExcel().then(base64data => {
    const binary = atob(base64data)
    const len = binary.length
    const bytes = new Uint8Array(len)
    for (let i = 0; i < len; i++) {
      bytes[i] = binary.charCodeAt(i)
    }
    const blob = new Blob([bytes], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = 'items.xlsx'
    link.click()
    setTimeout(() => URL.revokeObjectURL(link.href), 1000)
  }).catch(err => {
    alert('Ошибка экспорта: ' + err)
  })
}
function closeAddModal() {
  showAddModal.value = false
  newStock.value = { item_id: 0, warehouse_id: 0, quantity: 0 }
}
async function reloadWeeklyTrend() {
  weeklyStockData.value = await GetWeeklyStockTrend() || []
}

function confirmAddStock() {
  const filled = Object.values(newStock.value).every(v => v !== '' && v !== 0)
  if (!filled) {
    alert('Пожалуйста, заполните все поля')
    return
  }

  AddStock(
    newStock.value.item_id,
    newStock.value.quantity,
    newStock.value.warehouse_id
  ).then(() => {
    closeAddModal()
    reloadWeeklyTrend()
    refreshAllStockDetails()
    if (selectedWarehouseId.value === 0) {
      GetStockDetails().then(data => {
        stockList.value = normalizeStockRows(data)
      })
    } else {
      FindStockByWarehouse(selectedWarehouseId.value).then(data => {
        stockList.value = data.map(s => ({
          id: s.item_id,
          item_id: s.item_id,
          warehouse_id: s.warehouse_id,
          name: s.name,
          sku: s.sku,
          warehouse: s.warehouse, // ← теперь name уже приходит из бэка
          quantity: s.quantity
        }))
      })
    }
  }).catch(err => {
    alert("Ошибка при добавлении остатка")
    console.error(err)
  })
}

const newItem = ref({
  sku: "",
  name: "",
  description: "",
  uom: "",
  reorder_level: 0,
  reorder_qty: 0,
  price: 0,
  cost: 0,
  category: ""
})

function openAddItemModal() {
  Object.assign(newItem.value, {
    sku: "",
    name: "",
    description: "",
    uom: "",
    reorder_level: 0,
    reorder_qty: 0,
    price: 0,
    cost: 0,
    category: ""
  })
  showAddItemModal.value = true
}

function openEditItemModal(item) {
  itemToEdit.value = { ...item }
  showEditItemModal.value = true
}

async function confirmAddItem() {
  // простая валидация
  if (!newItem.value.sku || !newItem.value.name) {
    alert("Заполните все обязательные поля (артикул и наименование)");
    return;
  }
  try {
    await AddItem(newItem.value)
    showAddItemModal.value = false
    // обнови список
    items.value = await GetItems() || []
  } catch (e) {
    alert('Ошибка при добавлении: ' + (e?.message || ''))
  }
}

async function confirmEditItem() {
  if (!itemToEdit.value.sku || !itemToEdit.value.name) {
    alert("Заполните обязательные поля");
    return;
  }
  // Явно ставим null, если поля пустые (или 0? зависит от бизнес-логики)
  if (itemToEdit.value.price === "") itemToEdit.value.price = null;
  if (itemToEdit.value.cost === "") itemToEdit.value.cost = null;
  try {
    await UpdateItem(itemToEdit.value)
    showEditItemModal.value = false
    items.value = await GetItems() || []
  } catch (e) {
    alert('Ошибка при обновлении: ' + (e?.message || ''))
  }
}

async function deleteItem(item) {
  if (!confirm(`Удалить товар "${item.name}"?`)) return
  try {
    await RemoveItem(item.sku)
    items.value = await GetItems() || []
  } catch (e) {
    alert('Ошибка при удалении: ' + (e?.message || ''))
  }
}

function closeEditModal() {
  showEditModal.value = false
  stockToEdit.value = null
}

async function addWarehouse() {
  if (!newWarehouse.value.name.trim()) return
  await AddWarehouse(newWarehouse.value)
  await loadWarehouses()
  closeAddModal()
}

async function updateWarehouse() {
  if (!editWarehouseData.value.name.trim()) return
  await EditWarehouse(editWarehouseData.value)
  await loadWarehouses()
  closeEditModal()
}

function editWarehouse(w) {
  editWarehouseData.value = { ...w }
  showEditModal.value = true
}

const filteredWarehouses = computed(() =>
  warehouses.value.filter(w =>
    w.name.toLowerCase().includes(warehouseSearch.value.toLowerCase()) ||
    (w.location || '').toLowerCase().includes(warehouseSearch.value.toLowerCase())
  )
)

const dashboardStockRows = computed(() => {
  const warehouseId = Number(selectedRiskWarehouseId.value)
  return allStockList.value.filter(row => warehouseId === 0 || row.warehouse_id === warehouseId)
})

const lowStockItems = computed(() => {
  const stockMap = new Map()
  for (const row of dashboardStockRows.value) {
    const current = stockMap.get(row.item_id)
    if (!current) {
      stockMap.set(row.item_id, {
        warehouse_id: row.warehouse_id,
        warehouse: row.warehouse,
        quantity: Number(row.quantity) || 0
      })
      continue
    }
    current.quantity += Number(row.quantity) || 0
  }

  return items.value
    .map(item => {
      const stockInfo = stockMap.get(item.item_id)
      const currentStock = stockInfo?.quantity ?? 0
      const reorderLevel = Number(item.reorder_level) || 0
      const reorderQty = Number(item.reorder_qty) || 0
      const shortage = Math.max(reorderLevel - currentStock, 0)
      const recommendedOrder = shortage > 0
        ? Math.max(reorderQty, shortage)
        : 0

      if (reorderLevel <= 0 || currentStock > reorderLevel) {
        return null
      }

      const severity = currentStock === 0 ? 'critical' : currentStock <= Math.ceil(reorderLevel * 0.4) ? 'warning' : 'attention'

      return {
        item_id: item.item_id,
        warehouse_id: stockInfo?.warehouse_id || 0,
        warehouse: stockInfo?.warehouse || (selectedRiskWarehouseId.value === 0 ? 'Не распределен' : 'Нет остатков'),
        name: item.name,
        sku: item.sku,
        current_stock: currentStock,
        reorder_level: reorderLevel,
        recommended_order: recommendedOrder,
        estimated_cost: recommendedOrder * itemCost(item),
        severity,
        status_label: severity === 'critical' ? 'Критично' : severity === 'warning' ? 'Низкий остаток' : 'Нужен контроль'
      }
    })
    .filter(Boolean)
    .sort((a, b) => {
      if (a.severity !== b.severity) {
        return ['critical', 'warning', 'attention'].indexOf(a.severity) - ['critical', 'warning', 'attention'].indexOf(b.severity)
      }
      return a.current_stock - b.current_stock
    })
})

const criticalLowStockCount = computed(() =>
  lowStockItems.value.filter(item => item.severity === 'critical').length
)

const recommendedRestockUnits = computed(() =>
  lowStockItems.value.reduce((acc, item) => acc + item.recommended_order, 0)
)

const recommendedRestockCost = computed(() =>
  lowStockItems.value.reduce((acc, item) => acc + item.estimated_cost, 0)
)

const supplierByItem = computed(() => {
  const map = new Map()
  const rows = [...deliveriesList.value].sort((a, b) => new Date(b.date) - new Date(a.date))
  for (const row of rows) {
    if (!map.has(row.item_id)) {
      map.set(row.item_id, {
        supplier_id: row.supplier_id,
        supplier_name: row.supplier,
        last_inbound_date: row.date
      })
    }
  }
  return map
})

function itemOutboundStats(itemID, warehouseID) {
  const rows = outboundList.value.filter(row =>
    row.item_id === itemID &&
    (!warehouseID || row.warehouse_id === warehouseID)
  )
  const total = rows.reduce((acc, row) => acc + Number(row.quantity || 0), 0)
  return {
    total,
    last_date: rows.reduce((latest, row) => {
      const date = normalizeDateInput(row.date)
      return !latest || date > latest ? date : latest
    }, '')
  }
}

const purchasePlanRows = computed(() => {
  const warehouseId = Number(selectedRiskWarehouseId.value)
  const daysElapsed = Math.max(new Date().getDate(), 1)
  const rows = []

  for (const item of items.value) {
    const itemStocks = allStockList.value.filter(row =>
      row.item_id === item.item_id &&
      (warehouseId === 0 || row.warehouse_id === warehouseId)
    )
    const planningStocks = itemStocks.length
      ? itemStocks
      : [{
          item_id: item.item_id,
          warehouse_id: warehouseId || 0,
          warehouse: warehouseId ? warehouses.value.find(w => w.warehouse_id === warehouseId)?.name || 'Склад' : 'Без остатка',
          quantity: 0
        }]

    for (const stock of planningStocks) {
      const outboundStats = itemOutboundStats(item.item_id, stock.warehouse_id)
      const dailyDemand = outboundStats.total / daysElapsed
      const projectedDemand = Math.ceil(dailyDemand * Number(purchasePlanHorizon.value || 30))
      const currentStock = Number(stock.quantity || 0)
      const reorderLevel = Number(item.reorder_level) || 0
      const reorderQty = Number(item.reorder_qty) || 0
      const projectedStock = currentStock - projectedDemand
      const shortage = Math.max(reorderLevel - projectedStock, 0)
      const needsOrder = currentStock <= reorderLevel || projectedStock <= reorderLevel
      const recommendedOrder = needsOrder ? Math.max(reorderQty, shortage, 1) : 0
      const daysLeft = dailyDemand > 0 ? currentStock / dailyDemand : Infinity
      const supplier = supplierByItem.value.get(item.item_id) || {}
      let severity = 'stable'
      let statusLabel = 'В норме'

      if (currentStock === 0 || projectedStock <= 0) {
        severity = 'critical'
        statusLabel = 'Критично'
      } else if (reorderLevel > 0 && currentStock <= reorderLevel) {
        severity = 'warning'
        statusLabel = 'Низкий остаток'
      } else if (projectedStock <= reorderLevel) {
        severity = 'attention'
        statusLabel = 'Скоро закончится'
      }

      rows.push({
        item_id: item.item_id,
        warehouse_id: stock.warehouse_id,
        warehouse: stock.warehouse,
        name: item.name,
        sku: item.sku,
        current_stock: currentStock,
        reorder_level: reorderLevel,
        reorder_qty: reorderQty,
        daily_demand: dailyDemand,
        projected_stock: projectedStock,
        recommended_order: recommendedOrder,
        estimated_cost: recommendedOrder * itemCost(item),
        supplier_id: supplier.supplier_id || 0,
        supplier_name: supplier.supplier_name || '',
        last_inbound_date: supplier.last_inbound_date || '',
        last_outbound_date: outboundStats.last_date,
        days_left: daysLeft,
        days_left_label: Number.isFinite(daysLeft) ? `${Math.max(Math.floor(daysLeft), 0)} дн.` : 'Без расхода',
        severity,
        status_label: statusLabel
      })
    }
  }

  return rows.sort((a, b) => {
    const order = { critical: 0, warning: 1, attention: 2, stable: 3 }
    if (order[a.severity] !== order[b.severity]) return order[a.severity] - order[b.severity]
    if (b.recommended_order !== a.recommended_order) return b.recommended_order - a.recommended_order
    return a.name.localeCompare(b.name, 'ru')
  })
})

const filteredPurchasePlanItems = computed(() => {
  const search = purchasePlanSearch.value.toLowerCase()
  return purchasePlanRows.value.filter(item => {
    const matchesPriority =
      purchasePlanPriority.value === 'all' ||
      (purchasePlanPriority.value === 'order' && item.recommended_order > 0) ||
      item.severity === purchasePlanPriority.value
    const matchesSupplier = Number(purchasePlanSupplierId.value) === 0 || item.supplier_id === Number(purchasePlanSupplierId.value)

    return matchesPriority &&
      matchesSupplier &&
      (
        item.name.toLowerCase().includes(search) ||
        item.sku.toLowerCase().includes(search) ||
        item.warehouse.toLowerCase().includes(search) ||
        (item.supplier_name || '').toLowerCase().includes(search)
      )
  })
})

function purchaseRowKey(item) {
  return `${item.item_id}-${item.warehouse_id}`
}

function isPurchaseRowSelected(item) {
  return selectedPurchaseKeys.value.has(purchaseRowKey(item))
}

function togglePurchaseRow(item) {
  const next = new Set(selectedPurchaseKeys.value)
  const key = purchaseRowKey(item)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  selectedPurchaseKeys.value = next
}

function selectAllPurchaseRows() {
  selectedPurchaseKeys.value = new Set(
    filteredPurchasePlanItems.value
      .filter(item => item.recommended_order > 0)
      .map(purchaseRowKey)
  )
}

function clearPurchaseSelection() {
  selectedPurchaseKeys.value = new Set()
}

const selectedPurchasePlanItems = computed(() =>
  purchasePlanRows.value.filter(item => selectedPurchaseKeys.value.has(purchaseRowKey(item)))
)

const filteredPurchasePlanUnits = computed(() =>
  filteredPurchasePlanItems.value.reduce((acc, item) => acc + item.recommended_order, 0)
)

const filteredPurchasePlanCost = computed(() =>
  filteredPurchasePlanItems.value.reduce((acc, item) => acc + item.estimated_cost, 0)
)

const selectedPurchasePlanUnits = computed(() =>
  selectedPurchasePlanItems.value.reduce((acc, item) => acc + item.recommended_order, 0)
)

const selectedPurchasePlanCost = computed(() =>
  selectedPurchasePlanItems.value.reduce((acc, item) => acc + item.estimated_cost, 0)
)

const purchasePlanOrderCount = computed(() =>
  purchasePlanRows.value.filter(item => item.recommended_order > 0).length
)

const purchasePlanAttentionCount = computed(() =>
  purchasePlanRows.value.filter(item => ['critical', 'warning', 'attention'].includes(item.severity)).length
)

const topPurchaseRisk = computed(() =>
  purchasePlanRows.value.find(item => item.recommended_order > 0) || null
)

const purchasePrioritySummary = computed(() => [
  { key: 'all', label: 'Все', count: purchasePlanRows.value.length },
  { key: 'order', label: 'К закупке', count: purchasePlanRows.value.filter(item => item.recommended_order > 0).length },
  { key: 'critical', label: 'Критично', count: purchasePlanRows.value.filter(item => item.severity === 'critical').length },
  { key: 'warning', label: 'Низкий остаток', count: purchasePlanRows.value.filter(item => item.severity === 'warning').length },
  { key: 'attention', label: 'Скоро закончится', count: purchasePlanRows.value.filter(item => item.severity === 'attention').length },
  { key: 'stable', label: 'В норме', count: purchasePlanRows.value.filter(item => item.severity === 'stable').length }
])

function purchasePlanExportRows() {
  const rows = selectedPurchasePlanItems.value.length ? selectedPurchasePlanItems.value : filteredPurchasePlanItems.value
  return rows.map(item => [
    item.status_label,
    item.name,
    item.sku,
    item.supplier_name || '',
    item.warehouse,
    item.current_stock,
    item.reorder_level,
    formatNumber(item.daily_demand, 1),
    item.days_left_label,
    item.recommended_order,
    Math.round(item.estimated_cost)
  ])
}

function exportPurchasePlanCsv() {
  downloadCsv(
    'purchase_plan.csv',
    ['Приоритет', 'Товар', 'SKU', 'Поставщик', 'Склад', 'Остаток', 'Мин. остаток', 'Расход/день', 'Хватит на', 'Заказать', 'Бюджет'],
    purchasePlanExportRows()
  )
}

async function copyPurchasePlan() {
  const rows = purchasePlanExportRows()
  const text = [
    ['Приоритет', 'Товар', 'SKU', 'Поставщик', 'Склад', 'Остаток', 'Мин. остаток', 'Расход/день', 'Хватит на', 'Заказать', 'Бюджет'].join('\t'),
    ...rows.map(row => row.join('\t'))
  ].join('\n')
  try {
    await navigator.clipboard.writeText(text)
    alert('План закупок скопирован')
  } catch {
    exportPurchasePlanCsv()
  }
}

const warehouseRiskSummary = computed(() =>
  warehouses.value
    .map(warehouse => {
      const scoped = lowStockItems.value.filter(item => item.warehouse_id === warehouse.warehouse_id)
      return {
        warehouse_id: warehouse.warehouse_id,
        name: warehouse.name,
        low_count: scoped.length,
        critical_count: scoped.filter(item => item.severity === 'critical').length
      }
    })
    .filter(item => item.low_count > 0)
    .sort((a, b) => {
      if (b.critical_count !== a.critical_count) return b.critical_count - a.critical_count
      return b.low_count - a.low_count
    })
)

function warehouseStats(warehouseID) {
  const rows = allStockList.value.filter(row => row.warehouse_id === warehouseID)
  return {
    sku_count: new Set(rows.map(row => row.item_id)).size,
    units: rows.reduce((acc, row) => acc + Number(row.quantity || 0), 0),
    risk_count: lowStockItems.value.filter(item => item.warehouse_id === warehouseID).length
  }
}

const activeWarehouseCount = computed(() =>
  warehouses.value.filter(warehouse => warehouseStats(warehouse.warehouse_id).sku_count > 0).length
)

function supplierStats(supplierID) {
  const rows = deliveriesList.value.filter(row => row.supplier_id === supplierID)
  return {
    deliveries: rows.length,
    units: rows.reduce((acc, row) => acc + Number(row.quantity || 0), 0),
    last_date: rows.reduce((latest, row) => {
      const date = normalizeDateInput(row.date)
      return !latest || date > latest ? date : latest
    }, '')
  }
}

const activeSupplierCount = computed(() =>
  suppliers.value.filter(supplier => supplierStats(supplier.supplier_id).deliveries > 0).length
)

const supplierInboundUnits = computed(() =>
  deliveriesList.value.reduce((acc, row) => acc + Number(row.quantity || 0), 0)
)

function exportMovementsCsv() {
  downloadCsv(
    'movements.csv',
    ['Дата', 'Тип', 'Товар', 'Склад', 'Количество', 'Поставщик', 'Описание'],
    filteredMovements.value.map(row => [
      formatDate(row.date),
      movementTypeName(row.type),
      row.item_name,
      row.warehouse_name,
      row.quantity,
      row.supplier_name || '',
      row.destination || ''
    ])
  )
}

function formatMoney(value) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0
  }).format(Number(value) || 0)
}


onMounted(async () => {
  if (localStorage.getItem('loggedIn') === 'true' && !user.value) {
    loggedIn.value = false
    localStorage.removeItem('loggedIn')
    localStorage.removeItem('currentUser')
  }
  console.log('items:', items.value)
  GetItems().then(data => items.value = data || []);
  await GetWeeklyStockTrend().then(data => weeklyStockData.value = data)

  GetDashboard().then(data => {
    totalStock.value = data.total_stock
    itemCount.value = data.item_count
    monthlyOrders.value = data.monthly_orders
    newItems.value = data.new_items
  })
  GetSuppliers().then(data => suppliers.value = data || [])
  await loadUsers()
  GetStockDetails().then(data => {
    const rows = normalizeStockRows(data)
    stockList.value = rows
    allStockList.value = rows
  })

  await reloadMovements()
  await reloadTransfers()
  await reloadOutbound()
  GetInboundDetails().then(data => {
    deliveriesList.value = data || [];
  }).catch(err => {
    console.error("Ошибка загрузки поставок:", err)
  })
  GetTopItems().then(data => topItems.value = data)
  GetWarehouses().then(data => warehouses.value.push(...data))
  GetTurnoverByWarehouse().then(data => turnoverData.value = data)
})
watch([moveType, moveWarehouseId], reloadMovements)
watch(currentTab, async (tab) => {
  if (tab === 'Дашборд') {
    const data = await GetDashboard()
    totalStock.value = data.total_stock
    itemCount.value = data.item_count
    monthlyOrders.value = data.monthly_orders
    newItems.value = data.new_items
    topItems.value = await GetTopItems() || []
    turnoverData.value = await GetTurnoverByWarehouse() || []
    await refreshAllStockDetails()
  }
  if (tab === 'Движения') {
    await reloadMovements()
    await reloadTransfers()
  }
  if (tab === 'План закупок') {
    items.value = await GetItems() || []
    suppliers.value = await GetSuppliers() || []
    deliveriesList.value = await GetInboundDetails() || []
    await reloadOutbound()
    await refreshAllStockDetails()
  }
  if (tab === 'Поставщики') {
    suppliers.value = await GetSuppliers() || []
    deliveriesList.value = await GetInboundDetails() || []
  }
  if (tab === 'Поставки') {
    deliveriesList.value = await GetInboundDetails() || []
  }
  if (tab === 'Товары') {
    items.value = await GetItems() || []
    await refreshAllStockDetails()
  }
  if (tab === 'Склады') {
    await loadWarehouses()
    await refreshAllStockDetails()
  }
  if (tab === 'Отгрузки') {
    await reloadOutbound()
    await refreshAllStockDetails()
  }
  if (tab === 'Пользователи' && user.value?.role === 'admin') {
    try {
      users.value = await GetUsers() || []
    } catch (e) {
      users.value = []
    }
  }
})


watch(selectedWarehouseId, (id) => {
  const warehouseId = Number(id)
  if (warehouseId === 0) {
    GetStockDetails().then(data => {
      const rows = normalizeStockRows(data)
      stockList.value = rows
      allStockList.value = rows
    })
  } else {
    FindStockByWarehouse(warehouseId).then(data => {
      stockList.value = data.map(s => ({
        id: s.stock_id,
        stock_id: s.stock_id,
        item_id: s.item_id,
        warehouse_id: s.warehouse_id,
        name: s.name,
        sku: s.sku,
        warehouse: warehouses.value.find(w => w.warehouse_id === warehouseId)?.name || s.warehouse,
        quantity: s.quantity
      }))
    })
  }
})


function stockStatusFor(row) {
  const reorderLevel = Number(row.reorder_level) || 0
  const quantity = Number(row.quantity) || 0
  if (quantity === 0) return { key: 'zero', className: 'critical', label: 'Нет остатка' }
  if (reorderLevel > 0 && quantity <= reorderLevel) return { key: 'low', className: 'warning', label: 'Ниже минимума' }
  if (reorderLevel > 0 && quantity <= Math.ceil(reorderLevel * 1.5)) return { key: 'watch', className: 'attention', label: 'Под контролем' }
  return { key: 'ok', className: 'stable', label: 'В норме' }
}

const enrichedStockList = computed(() =>
  stockList.value.map(row => {
    const item = items.value.find(i => i.item_id === row.item_id) || {}
    const status = stockStatusFor({
      ...row,
      reorder_level: item.reorder_level
    })
    return {
      ...row,
      reorder_level: Number(item.reorder_level) || 0,
      reorder_qty: Number(item.reorder_qty) || 0,
      stock_status_key: status.key,
      stock_status: status.className,
      status_label: status.label
    }
  })
)

const filteredStockList = computed(() => {
  const search = searchQuery.value.toLowerCase()
  return enrichedStockList.value.filter(item =>
    (stockStatusFilter.value === 'all' || item.stock_status_key === stockStatusFilter.value) &&
    (
      (item.name || '').toLowerCase().includes(search) ||
      (item.sku || '').toLowerCase().includes(search) ||
      (item.warehouse || '').toLowerCase().includes(search)
    )
  )
})

const filteredStockUnits = computed(() =>
  filteredStockList.value.reduce((acc, item) => acc + Number(item.quantity || 0), 0)
)

const lowStockRowCount = computed(() =>
  enrichedStockList.value.filter(item => ['zero', 'low'].includes(item.stock_status_key)).length
)

const zeroStockRowCount = computed(() =>
  enrichedStockList.value.filter(item => item.stock_status_key === 'zero').length
)
</script>

<style scoped>
.dashboard-risk-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: 1.2rem;
  margin-top: 1.2rem;
}

.dashboard-risk-card,
.dashboard-risk-sidecard {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(247, 249, 252, 0.98));
}

.dashboard-risk-header {
  align-items: flex-start;
  gap: 1rem;
}

.dashboard-subtitle,
.dashboard-row-meta {
  color: #6b7280;
  font-size: 0.85rem;
}

.dashboard-risk-filter {
  min-width: 220px;
  max-width: 320px;
}

.risk-summary-cards {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.8rem;
  margin-bottom: 1rem;
}

.risk-summary-card {
  border-radius: 16px;
  padding: 0.9rem 1rem;
  border: 1px solid rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.risk-summary-card strong {
  font-size: 1.15rem;
  color: #111827;
}

.risk-summary-card.warning { background: #fff7ed; }
.risk-summary-card.danger { background: #fef2f2; }
.risk-summary-card.neutral { background: #f8fafc; }
.risk-summary-card.accent { background: #eff6ff; }

.risk-summary-label {
  color: #4b5563;
  font-size: 0.84rem;
}

.stock-health-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.32rem 0.7rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.01em;
}

.stock-health-badge.critical {
  background: #fee2e2;
  color: #b91c1c;
}

.stock-health-badge.warning {
  background: #ffedd5;
  color: #c2410c;
}

.stock-health-badge.attention {
  background: #fef9c3;
  color: #a16207;
}

.stock-health-badge.stable {
  background: #dcfce7;
  color: #166534;
}

.plan-workbench {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(240px, 0.7fr);
  gap: 1rem;
  margin-bottom: 1.2rem;
}

.plan-workbench-main,
.plan-workbench-side {
  background: #fff;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 14px;
  padding: 1rem 1.2rem;
  box-shadow: 0 4px 18px rgba(219, 234, 254, 0.18);
}

.plan-workbench-side {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.plan-workbench-side strong {
  color: #111827;
  font-size: 1.08rem;
}

.plan-actions {
  display: flex;
  gap: 0.55rem;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.plan-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.plan-chip {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 999px;
  background: #f8fafc;
  color: #1f2937;
  padding: 0.38rem 0.7rem;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-weight: 700;
  cursor: pointer;
}

.plan-chip.active {
  background: #eff6ff;
  border-color: #93c5fd;
  color: #1d4ed8;
}

.plan-chip strong {
  font-size: 0.86rem;
}

button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.warehouse-risk-list {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.warehouse-risk-row {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  padding: 0.9rem 1rem;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px solid rgba(15, 23, 42, 0.08);
}

.warehouse-risk-name {
  font-weight: 700;
  color: #111827;
}

.warehouse-risk-values {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  color: #6b7280;
  font-size: 0.82rem;
}

.warehouse-risk-values strong {
  color: #b91c1c;
  font-size: 1.1rem;
}

@media (max-width: 1100px) {
  .dashboard-risk-grid {
    grid-template-columns: 1fr;
  }

  .risk-summary-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .plan-workbench {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .dashboard-risk-filter {
    min-width: 100%;
    max-width: 100%;
  }

  .risk-summary-cards {
    grid-template-columns: 1fr;
  }
}
</style>
