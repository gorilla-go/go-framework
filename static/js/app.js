// 主要应用脚本
document.addEventListener('DOMContentLoaded', function() {
    console.log('应用已加载');
    
    // 绑定所有按钮点击事件
    bindButtonEvents();
    
    // 添加表单验证
    setupFormValidation();
});

// 绑定按钮点击事件
function bindButtonEvents() {
    const buttons = document.querySelectorAll('.btn');
    
    buttons.forEach(button => {
        button.addEventListener('click', function(event) {
            console.log('按钮被点击:', this.textContent);
        });
    });
}

// 设置表单验证
function setupFormValidation() {
    const forms = document.querySelectorAll('form');
    
    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            if (!validateForm(this)) {
                event.preventDefault();
                console.log('表单验证失败');
            } else {
                console.log('表单验证成功，提交中...');
            }
        });
    });
}

// 表单验证
function validateForm(form) {
    let isValid = true;
    
    // 验证必填字段
    const requiredFields = form.querySelectorAll('[required]');
    requiredFields.forEach(field => {
        if (!field.value.trim()) {
            isValid = false;
            field.classList.add('is-invalid');
            
            // 添加错误消息
            const errorMsg = document.createElement('div');
            errorMsg.className = 'invalid-feedback';
            errorMsg.textContent = '此字段为必填项';
            
            // 移除可能存在的旧错误消息
            const existingError = field.parentNode.querySelector('.invalid-feedback');
            if (existingError) {
                existingError.remove();
            }
            
            field.parentNode.appendChild(errorMsg);
        } else {
            field.classList.remove('is-invalid');
            const existingError = field.parentNode.querySelector('.invalid-feedback');
            if (existingError) {
                existingError.remove();
            }
        }
    });
    
    return isValid;
}

// 发送API请求的辅助函数
async function apiRequest(url, method = 'GET', data = null) {
    try {
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest'
            }
        };
        
        if (data && (method === 'POST' || method === 'PUT' || method === 'PATCH')) {
            options.body = JSON.stringify(data);
        }
        
        // 添加认证令牌（如果存在）
        const token = localStorage.getItem('auth_token');
        if (token) {
            options.headers['Authorization'] = `Bearer ${token}`;
        }
        
        const response = await fetch(url, options);
        const responseData = await response.json();
        
        if (!response.ok) {
            throw new Error(responseData.message || '请求失败');
        }
        
        return responseData;
    } catch (error) {
        console.error('API请求错误:', error);
        throw error;
    }
}

// 显示通知消息
function showNotification(message, type = 'success') {
    const notification = document.createElement('div');
    notification.className = `alert alert-${type}`;
    notification.textContent = message;
    
    // 添加关闭按钮
    const closeBtn = document.createElement('button');
    closeBtn.className = 'close';
    closeBtn.innerHTML = '&times;';
    closeBtn.addEventListener('click', function() {
        notification.remove();
    });
    
    notification.appendChild(closeBtn);
    
    // 添加到页面
    const container = document.querySelector('.container') || document.body;
    container.prepend(notification);
    
    // 5秒后自动消失
    setTimeout(() => {
        notification.remove();
    }, 5000);
} 