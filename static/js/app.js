// 主应用程序类
class ProxyTestApp {
    constructor() {
        this.apiBase = '';
        this.wsConnections = new Map();
        this.testResults = [];
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadSystemInfo();
        this.initializeComponents();
    }

    setupEventListeners() {
        // HTTP测试按钮事件
        document.getElementById('testHttpBtn')?.addEventListener('click', () => this.testHttpRequest());
        document.getElementById('testAllMethodsBtn')?.addEventListener('click', () => this.testAllMethods());
        document.getElementById('testConcurrentBtn')?.addEventListener('click', () => this.testConcurrent());
        document.getElementById('testStressBtn')?.addEventListener('click', () => this.testStress());
        
        // WebSocket测试按钮事件
        document.getElementById('connectWsBtn')?.addEventListener('click', () => this.connectWebSocket());
        document.getElementById('disconnectWsBtn')?.addEventListener('click', () => this.disconnectWebSocket());
        document.getElementById('sendWsMessageBtn')?.addEventListener('click', () => this.sendWebSocketMessage());
        
        // 工具按钮事件
        document.getElementById('clearResultsBtn')?.addEventListener('click', () => this.clearResults());
        document.getElementById('exportResultsBtn')?.addEventListener('click', () => this.exportResults());
        document.getElementById('loadSystemInfoBtn')?.addEventListener('click', () => this.loadSystemInfo());
        
        // 表单提交事件
        document.getElementById('httpTestForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.testHttpRequest();
        });
        
        document.getElementById('wsTestForm')?.addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendWebSocketMessage();
        });
    }

    initializeComponents() {
        // 初始化工具提示
        this.initTooltips();
        
        // 初始化代码编辑器
        this.initCodeEditors();
        
        // 初始化统计图表
        this.initCharts();
        
        // 加载预设配置
        this.loadPresets();
    }

    initTooltips() {
        // 为所有带有data-tooltip属性的元素添加工具提示
        document.querySelectorAll('[data-tooltip]').forEach(element => {
            element.classList.add('tooltip');
        });
    }

    initCodeEditors() {
        // 初始化JSON编辑器
        const jsonEditor = document.getElementById('requestBody');
        if (jsonEditor) {
            jsonEditor.addEventListener('input', (e) => {
                this.validateJSON(e.target.value);
            });
        }
    }

    initCharts() {
        // 初始化性能图表
        this.performanceChart = null;
        this.requestChart = null;
    }

    loadPresets() {
        // 加载预设的测试配置
        const presets = {
            'basic-get': {
                method: 'GET',
                url: '/api/test',
                headers: {},
                body: ''
            },
            'json-post': {
                method: 'POST',
                url: '/api/test',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ message: 'Hello, World!' })
            },
            'file-upload': {
                method: 'POST',
                url: '/api/upload',
                headers: {},
                body: ''
            }
        };
        
        this.populatePresetSelect(presets);
    }

    populatePresetSelect(presets) {
        const select = document.getElementById('presetSelect');
        if (select) {
            Object.keys(presets).forEach(key => {
                const option = document.createElement('option');
                option.value = key;
                option.textContent = key.replace('-', ' ').toUpperCase();
                select.appendChild(option);
            });
            
            select.addEventListener('change', (e) => {
                if (e.target.value && presets[e.target.value]) {
                    this.loadPreset(presets[e.target.value]);
                }
            });
        }
    }

    loadPreset(preset) {
        document.getElementById('requestMethod').value = preset.method;
        document.getElementById('requestUrl').value = preset.url;
        document.getElementById('requestHeaders').value = JSON.stringify(preset.headers, null, 2);
        document.getElementById('requestBody').value = preset.body;
    }

    async testHttpRequest() {
        const method = document.getElementById('requestMethod').value;
        const url = document.getElementById('requestUrl').value;
        const headers = this.parseJSON(document.getElementById('requestHeaders').value) || {};
        const body = document.getElementById('requestBody').value;
        
        if (!url) {
            this.showAlert('请输入请求URL', 'warning');
            return;
        }

        this.showLoading('testHttpBtn');
        const startTime = performance.now();
        
        try {
            const config = {
                method: method,
                headers: headers
            };
            
            if (body && ['POST', 'PUT', 'PATCH'].includes(method)) {
                config.body = body;
            }
            
            const response = await fetch(url, config);
            const responseTime = performance.now() - startTime;
            
            const result = {
                id: Date.now(),
                timestamp: new Date().toISOString(),
                request: {
                    method,
                    url,
                    headers,
                    body: body || null
                },
                response: {
                    status: response.status,
                    statusText: response.statusText,
                    headers: Object.fromEntries(response.headers.entries()),
                    body: await response.text(),
                    responseTime: Math.round(responseTime)
                },
                success: response.ok
            };
            
            this.addTestResult(result);
            this.displayResult(result);
            this.showAlert(`请求完成，耗时: ${result.response.responseTime}ms`, 'success');
            
        } catch (error) {
            const errorResult = {
                id: Date.now(),
                timestamp: new Date().toISOString(),
                request: { method, url, headers, body },
                error: error.message,
                success: false
            };
            
            this.addTestResult(errorResult);
            this.displayResult(errorResult);
            this.showAlert('请求失败: ' + error.message, 'danger');
        } finally {
            this.hideLoading('testHttpBtn');
        }
    }

    async testAllMethods() {
        const url = document.getElementById('requestUrl').value || '/api/test';
        const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS'];
        
        this.showLoading('testAllMethodsBtn');
        this.clearResults();
        
        for (const method of methods) {
            try {
                const config = {
                    method: method,
                    headers: { 'Content-Type': 'application/json' }
                };
                
                if (['POST', 'PUT', 'PATCH'].includes(method)) {
                    config.body = JSON.stringify({ test: true, method: method });
                }
                
                const startTime = performance.now();
                const response = await fetch(url, config);
                const responseTime = performance.now() - startTime;
                
                const result = {
                    id: Date.now() + Math.random(),
                    timestamp: new Date().toISOString(),
                    request: {
                        method,
                        url,
                        headers: config.headers,
                        body: config.body || null
                    },
                    response: {
                        status: response.status,
                        statusText: response.statusText,
                        headers: Object.fromEntries(response.headers.entries()),
                        body: method !== 'HEAD' ? await response.text() : '',
                        responseTime: Math.round(responseTime)
                    },
                    success: response.ok
                };
                
                this.addTestResult(result);
                this.displayResult(result);
                
            } catch (error) {
                const errorResult = {
                    id: Date.now() + Math.random(),
                    timestamp: new Date().toISOString(),
                    request: { method, url },
                    error: error.message,
                    success: false
                };
                
                this.addTestResult(errorResult);
                this.displayResult(errorResult);
            }
        }
        
        this.hideLoading('testAllMethodsBtn');
        this.showAlert(`完成所有HTTP方法测试`, 'success');
    }

    async testConcurrent() {
        const concurrency = parseInt(document.getElementById('concurrency').value) || 10;
        const requests = parseInt(document.getElementById('requests').value) || 100;
        const delay = parseInt(document.getElementById('delay').value) || 0;
        
        this.showLoading('testConcurrentBtn');
        
        try {
            const url = `/test/concurrent?concurrency=${concurrency}&requests=${requests}&delay=${delay}`;
            const response = await fetch(url);
            const result = await response.json();
            
            this.displayConcurrentResult(result);
            this.showAlert('并发测试完成', 'success');
            
        } catch (error) {
            this.showAlert('并发测试失败: ' + error.message, 'danger');
        } finally {
            this.hideLoading('testConcurrentBtn');
        }
    }

    async testStress() {
        const duration = parseInt(document.getElementById('stressDuration').value) || 60;
        const concurrency = parseInt(document.getElementById('stressConcurrency').value) || 20;
        
        this.showLoading('testStressBtn');
        
        try {
            const url = `/test/stress?duration=${duration}&concurrency=${concurrency}`;
            const response = await fetch(url);
            const result = await response.json();
            
            this.displayStressResult(result);
            this.showAlert('压力测试完成', 'success');
            
        } catch (error) {
            this.showAlert('压力测试失败: ' + error.message, 'danger');
        } finally {
            this.hideLoading('testStressBtn');
        }
    }

    connectWebSocket() {
        const wsUrl = document.getElementById('wsUrl').value || 'ws://localhost:8080/ws/connect';
        const wsType = document.getElementById('wsType').value || 'connect';
        
        if (this.wsConnections.has(wsType)) {
            this.showAlert('WebSocket已连接', 'warning');
            return;
        }
        
        try {
            const ws = new WebSocket(wsUrl.replace('/connect', '/' + wsType));
            
            ws.onopen = () => {
                this.wsConnections.set(wsType, ws);
                this.updateWebSocketStatus(wsType, 'connected');
                this.showAlert('WebSocket连接成功', 'success');
                this.addWebSocketMessage('系统', '连接已建立', 'info');
            };
            
            ws.onmessage = (event) => {
                let data;
                try {
                    data = JSON.parse(event.data);
                } catch {
                    data = event.data;
                }
                this.addWebSocketMessage('服务器', data, 'received');
            };
            
            ws.onclose = () => {
                this.wsConnections.delete(wsType);
                this.updateWebSocketStatus(wsType, 'disconnected');
                this.showAlert('WebSocket连接关闭', 'info');
                this.addWebSocketMessage('系统', '连接已关闭', 'warning');
            };
            
            ws.onerror = (error) => {
                this.showAlert('WebSocket错误: ' + error.message, 'danger');
                this.addWebSocketMessage('系统', '连接错误: ' + error.message, 'error');
            };
            
        } catch (error) {
            this.showAlert('WebSocket连接失败: ' + error.message, 'danger');
        }
    }

    disconnectWebSocket() {
        const wsType = document.getElementById('wsType').value || 'connect';
        const ws = this.wsConnections.get(wsType);
        
        if (ws) {
            ws.close();
            this.wsConnections.delete(wsType);
            this.updateWebSocketStatus(wsType, 'disconnected');
            this.showAlert('WebSocket已断开', 'info');
        } else {
            this.showAlert('没有活动的WebSocket连接', 'warning');
        }
    }

    sendWebSocketMessage() {
        const wsType = document.getElementById('wsType').value || 'connect';
        const message = document.getElementById('wsMessage').value;
        const ws = this.wsConnections.get(wsType);
        
        if (!ws) {
            this.showAlert('请先连接WebSocket', 'warning');
            return;
        }
        
        if (!message) {
            this.showAlert('请输入消息内容', 'warning');
            return;
        }
        
        try {
            const messageData = {
                type: 'message',
                data: message,
                timestamp: Date.now()
            };
            
            ws.send(JSON.stringify(messageData));
            this.addWebSocketMessage('客户端', messageData, 'sent');
            document.getElementById('wsMessage').value = '';
            
        } catch (error) {
            this.showAlert('发送消息失败: ' + error.message, 'danger');
        }
    }

    updateWebSocketStatus(type, status) {
        const indicator = document.getElementById('wsStatus');
        if (indicator) {
            indicator.className = `status-indicator status-${status === 'connected' ? 'success' : 'error'}`;
        }
        
        const statusText = document.getElementById('wsStatusText');
        if (statusText) {
            statusText.textContent = status === 'connected' ? '已连接' : '未连接';
        }
        
        // 更新按钮状态
        const connectBtn = document.getElementById('connectWsBtn');
        const disconnectBtn = document.getElementById('disconnectWsBtn');
        
        if (connectBtn && disconnectBtn) {
            if (status === 'connected') {
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
            } else {
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
            }
        }
    }

    addWebSocketMessage(sender, message, type) {
        const container = document.getElementById('wsMessages');
        if (!container) return;
        
        const messageElement = document.createElement('div');
        messageElement.className = `websocket-message ${type}`;
        
        const timestamp = new Date().toLocaleTimeString();
        const messageText = typeof message === 'object' ? JSON.stringify(message, null, 2) : message;
        
        messageElement.innerHTML = `
            <strong>${sender}</strong> <small class="text-muted">${timestamp}</small>
            <pre>${messageText}</pre>
        `;
        
        container.appendChild(messageElement);
        container.scrollTop = container.scrollHeight;
    }

    addTestResult(result) {
        this.testResults.push(result);
        this.updateTestStats();
    }

    displayResult(result) {
        const container = document.getElementById('testResults');
        if (!container) return;
        
        const resultElement = document.createElement('div');
        resultElement.className = 'card mb-3';
        
        const statusClass = result.success ? 'success' : 'danger';
        const methodClass = result.request.method ? result.request.method.toLowerCase() : 'unknown';
        
        resultElement.innerHTML = `
            <div class="card-header">
                <div class="d-flex justify-content-between align-items-center">
                    <div>
                        <span class="method-badge method-${methodClass}">${result.request.method}</span>
                        <span class="ms-2">${result.request.url}</span>
                    </div>
                    <div>
                        <span class="status-indicator status-${result.success ? 'success' : 'error'}"></span>
                        <small class="text-muted">${result.timestamp}</small>
                    </div>
                </div>
            </div>
            <div class="card-body">
                ${result.response ? `
                    <div class="row">
                        <div class="col-md-6">
                            <strong>状态:</strong> ${result.response.status} ${result.response.statusText}<br>
                            <strong>响应时间:</strong> ${result.response.responseTime}ms
                        </div>
                        <div class="col-md-6">
                            <strong>响应大小:</strong> ${result.response.body.length} bytes
                        </div>
                    </div>
                    <div class="mt-3">
                        <div class="json-viewer">${this.formatResponse(result.response.body)}</div>
                    </div>
                ` : `
                    <div class="alert alert-danger">
                        <strong>错误:</strong> ${result.error}
                    </div>
                `}
            </div>
        `;
        
        container.insertBefore(resultElement, container.firstChild);
    }

    displayConcurrentResult(result) {
        const container = document.getElementById('concurrentResults');
        if (!container) return;
        
        const data = result.data;
        container.innerHTML = `
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">${data.total_requests}</div>
                    <div class="stat-label">总请求数</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.success_requests}</div>
                    <div class="stat-label">成功请求</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.average_response_ms}ms</div>
                    <div class="stat-label">平均响应时间</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.requests_per_second.toFixed(2)}</div>
                    <div class="stat-label">每秒请求数</div>
                </div>
            </div>
            <div class="row">
                <div class="col-md-6">
                    <strong>最大响应时间:</strong> ${data.max_response_ms}ms<br>
                    <strong>最小响应时间:</strong> ${data.min_response_ms}ms<br>
                    <strong>失败请求:</strong> ${data.failed_requests}
                </div>
                <div class="col-md-6">
                    <strong>测试时长:</strong> ${data.duration_ms}ms<br>
                    <strong>开始时间:</strong> ${new Date(data.start_time * 1000).toLocaleString()}<br>
                    <strong>结束时间:</strong> ${new Date(data.end_time * 1000).toLocaleString()}
                </div>
            </div>
        `;
    }

    displayStressResult(result) {
        const container = document.getElementById('stressResults');
        if (!container) return;
        
        const data = result.data;
        container.innerHTML = `
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-number">${data.total_requests}</div>
                    <div class="stat-label">总请求数</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.success_requests}</div>
                    <div class="stat-label">成功请求</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.average_response_ms}ms</div>
                    <div class="stat-label">平均响应时间</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">${data.requests_per_second.toFixed(2)}</div>
                    <div class="stat-label">每秒请求数</div>
                </div>
            </div>
        `;
    }

    updateTestStats() {
        const totalTests = this.testResults.length;
        const successTests = this.testResults.filter(r => r.success).length;
        const failedTests = totalTests - successTests;
        
        const successRate = totalTests > 0 ? (successTests / totalTests * 100).toFixed(1) : 0;
        
        const statsContainer = document.getElementById('testStats');
        if (statsContainer) {
            statsContainer.innerHTML = `
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-number">${totalTests}</div>
                        <div class="stat-label">总测试数</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${successTests}</div>
                        <div class="stat-label">成功测试</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${failedTests}</div>
                        <div class="stat-label">失败测试</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${successRate}%</div>
                        <div class="stat-label">成功率</div>
                    </div>
                </div>
            `;
        }
    }

    async loadSystemInfo() {
        try {
            const response = await fetch('/test/system');
            const result = await response.json();
            
            this.displaySystemInfo(result.data);
            
        } catch (error) {
            this.showAlert('加载系统信息失败: ' + error.message, 'danger');
        }
    }

    async testSpecific(type) {
        const typeMap = {
            'memory': '内存',
            'cpu': 'CPU',
            'network': '网络',
            'fileio': '文件IO'
        };
        
        const typeName = typeMap[type] || type;
        this.showAlert(`开始${typeName}测试...`, 'info');
        
        try {
            let url = `/test/${type}`;
            
            // 为不同类型的测试设置默认参数
            switch(type) {
                case 'memory':
                    url += '?size=100&duration=10';
                    break;
                case 'cpu':
                    url += '?duration=10&cores=4';
                    break;
                case 'network':
                    url += '?size=1024';
                    break;
                case 'fileio':
                    url += '?operations=100';
                    break;
            }
            
            const response = await fetch(url);
            const result = await response.json();
            
            this.displaySpecificResult(type, result);
            this.showAlert(`${typeName}测试完成`, 'success');
            
        } catch (error) {
            this.showAlert(`${typeName}测试失败: ` + error.message, 'danger');
        }
    }

    displaySpecificResult(type, result) {
        const container = document.getElementById('specificResults');
        if (!container) return;
        
        const typeMap = {
            'memory': '内存测试',
            'cpu': 'CPU测试',
            'network': '网络测试',
            'fileio': '文件IO测试'
        };
        
        const typeName = typeMap[type] || type;
        const data = result.data;
        
        let resultHtml = `
            <div class="card mt-3">
                <div class="card-header">
                    <h6 class="mb-0">${typeName}结果</h6>
                </div>
                <div class="card-body">
        `;
        
        if (type === 'memory') {
            resultHtml += `
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-number">${data.allocated_mb || 0}MB</div>
                        <div class="stat-label">分配内存</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${data.operations || 0}</div>
                        <div class="stat-label">操作次数</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${data.duration_ms || 0}ms</div>
                        <div class="stat-label">测试时长</div>
                    </div>
                </div>
            `;
        } else if (type === 'cpu') {
            resultHtml += `
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-number">${data.cpu_usage || 0}%</div>
                        <div class="stat-label">CPU使用率</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${data.operations || 0}</div>
                        <div class="stat-label">运算次数</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-number">${data.duration_ms || 0}ms</div>
                        <div class="stat-label">测试时长</div>
                    </div>
                </div>
            `;
        } else {
            // 通用结果显示
            resultHtml += `
                <div class="json-viewer">${JSON.stringify(data, null, 2)}</div>
            `;
        }
        
        resultHtml += `
                </div>
            </div>
        `;
        
        container.innerHTML = resultHtml;
    }

    displaySystemInfo(data) {
        const container = document.getElementById('systemInfo');
        if (!container) return;
        
        container.innerHTML = `
            <div class="row">
                <div class="col-md-6">
                    <h6>系统信息</h6>
                    <p><strong>Go版本:</strong> ${data.go_version}</p>
                    <p><strong>CPU核心数:</strong> ${data.cpu_count}</p>
                    <p><strong>协程数:</strong> ${data.goroutine_count}</p>
                </div>
                <div class="col-md-6">
                    <h6>内存信息</h6>
                    <p><strong>已分配内存:</strong> ${data.memory.alloc_mb} MB</p>
                    <p><strong>系统内存:</strong> ${data.memory.sys_mb} MB</p>
                    <p><strong>堆对象数:</strong> ${data.memory.heap_objects}</p>
                </div>
            </div>
            <div class="row mt-3">
                <div class="col-md-6">
                    <h6>垃圾回收</h6>
                    <p><strong>GC次数:</strong> ${data.gc.num_gc}</p>
                    <p><strong>GC总暂停时间:</strong> ${(data.gc.pause_total_ns / 1000000).toFixed(2)} ms</p>
                </div>
                <div class="col-md-6">
                    <h6>其他信息</h6>
                    <p><strong>更新时间:</strong> ${new Date(data.timestamp * 1000).toLocaleString()}</p>
                </div>
            </div>
        `;
    }

    clearResults() {
        this.testResults = [];
        
        const containers = ['testResults', 'wsMessages', 'concurrentResults', 'stressResults'];
        containers.forEach(id => {
            const container = document.getElementById(id);
            if (container) {
                container.innerHTML = '';
            }
        });
        
        this.updateTestStats();
        this.showAlert('结果已清空', 'info');
    }

    exportResults() {
        if (this.testResults.length === 0) {
            this.showAlert('没有测试结果可导出', 'warning');
            return;
        }
        
        const data = {
            export_time: new Date().toISOString(),
            total_tests: this.testResults.length,
            results: this.testResults
        };
        
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = `proxy-test-results-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        
        URL.revokeObjectURL(url);
        this.showAlert('测试结果已导出', 'success');
    }

    formatResponse(response) {
        try {
            const parsed = JSON.parse(response);
            return JSON.stringify(parsed, null, 2);
        } catch {
            return response;
        }
    }

    parseJSON(str) {
        try {
            return JSON.parse(str);
        } catch {
            return null;
        }
    }

    validateJSON(str) {
        const indicator = document.getElementById('jsonValidator');
        if (!indicator) return;
        
        if (!str.trim()) {
            indicator.innerHTML = '';
            return;
        }
        
        try {
            JSON.parse(str);
            indicator.innerHTML = '<small class="text-success">✓ 有效的JSON</small>';
        } catch (error) {
            indicator.innerHTML = '<small class="text-danger">✗ 无效的JSON: ' + error.message + '</small>';
        }
    }

    showAlert(message, type = 'info') {
        // 创建或获取提醒容器
        let alertContainer = document.getElementById('alertContainer');
        if (!alertContainer) {
            alertContainer = document.createElement('div');
            alertContainer.id = 'alertContainer';
            alertContainer.className = 'position-fixed top-0 end-0 p-3';
            alertContainer.style.zIndex = '9999';
            document.body.appendChild(alertContainer);
        }
        
        const alert = document.createElement('div');
        alert.className = `alert alert-${type} alert-dismissible fade show`;
        alert.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        alertContainer.appendChild(alert);
        
        // 自动移除提醒
        setTimeout(() => {
            if (alert.parentNode) {
                alert.parentNode.removeChild(alert);
            }
        }, 5000);
    }

    showLoading(buttonId) {
        const button = document.getElementById(buttonId);
        if (button) {
            button.disabled = true;
            button.classList.add('loading');
            const originalText = button.textContent;
            button.setAttribute('data-original-text', originalText);
            button.innerHTML = '<span class="spinner-border spinner-border-sm me-2"></span>加载中...';
        }
    }

    hideLoading(buttonId) {
        const button = document.getElementById(buttonId);
        if (button) {
            button.disabled = false;
            button.classList.remove('loading');
            const originalText = button.getAttribute('data-original-text');
            if (originalText) {
                button.textContent = originalText;
            }
        }
    }
}

// 工具函数
function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDuration(milliseconds) {
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else if (minutes > 0) {
        return `${minutes}m ${seconds % 60}s`;
    } else {
        return `${seconds}s`;
    }
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        // 显示复制成功提示
        const toast = document.createElement('div');
        toast.className = 'toast position-fixed top-0 end-0 m-3';
        toast.innerHTML = `
            <div class="toast-body">
                已复制到剪贴板
            </div>
        `;
        document.body.appendChild(toast);
        
        setTimeout(() => {
            document.body.removeChild(toast);
        }, 2000);
    });
}

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    window.app = new ProxyTestApp();
}); 