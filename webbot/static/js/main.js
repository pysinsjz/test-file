// 全局 JavaScript 功能

// 页面加载完成后执行
$(document).ready(function() {
    // 初始化提示工具
    initTooltips();

    // 初始化通用功能
    initCommonFeatures();

    // 初始化页面特定功能
    initPageSpecificFeatures();
});

// 初始化提示工具
function initTooltips() {
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    var tooltipList = tooltipTriggerList.map(function(tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
}

// 初始化通用功能
function initCommonFeatures() {
    // 平滑滚动
    $('a[href^="#"]').on('click', function(e) {
        e.preventDefault();
        var target = $($(this).attr('href'));
        if (target.length) {
            $('html, body').animate({
                scrollTop: target.offset().top - 70
            }, 500);
        }
    });

    // 返回顶部按钮
    addBackToTopButton();

    // 表单验证增强
    enhanceFormValidation();

    // 复制到剪贴板功能
    initCopyToClipboard();
}

// 返回顶部按钮
function addBackToTopButton() {
    // 创建返回顶部按钮
    if (!$('#backToTop').length) {
        $('body').append(`
            <button id="backToTop" class="btn btn-primary position-fixed"
                    style="bottom: 20px; right: 20px; z-index: 1050; display: none; border-radius: 50%; width: 50px; height: 50px;">
                <i class="fas fa-arrow-up"></i>
            </button>
        `);
    }

    // 滚动事件处理
    $(window).scroll(function() {
        if ($(this).scrollTop() > 300) {
            $('#backToTop').fadeIn();
        } else {
            $('#backToTop').fadeOut();
        }
    });

    // 点击返回顶部
    $('#backToTop').click(function() {
        $('html, body').animate({scrollTop: 0}, 500);
    });
}

// 表单验证增强
function enhanceFormValidation() {
    // Bootstrap 表单验证
    $('.needs-validation').on('submit', function(e) {
        if (!this.checkValidity()) {
            e.preventDefault();
            e.stopPropagation();
        }
        $(this).addClass('was-validated');
    });

    // 实时验证
    $('.form-control').on('input', function() {
        if (this.checkValidity()) {
            $(this).removeClass('is-invalid').addClass('is-valid');
        } else {
            $(this).removeClass('is-valid').addClass('is-invalid');
        }
    });
}

// 复制到剪贴板功能
function initCopyToClipboard() {
    $('[data-copy]').click(function() {
        const text = $(this).data('copy');
        navigator.clipboard.writeText(text).then(function() {
            showToast('已复制到剪贴板', 'success');
        }).catch(function() {
            showToast('复制失败', 'error');
        });
    });
}

// 初始化页面特定功能
function initPageSpecificFeatures() {
    const currentPage = $('body').data('page');

    switch(currentPage) {
        case 'index':
            initIndexPage();
            break;
        case 'upload':
            initUploadPage();
            break;
        case 'result':
            initResultPage();
            break;
        case 'help':
            initHelpPage();
            break;
    }
}

// 首页初始化
function initIndexPage() {
    // 功能卡片动画
    $('.function-card').each(function(index) {
        $(this).css('animation-delay', (index * 0.1) + 's')
               .addClass('animate__animated animate__fadeInUp');
    });

    // 统计数字动画
    animateNumbers();
}

// 统计数字动画
function animateNumbers() {
    $('.stats-row h3').each(function() {
        const $this = $(this);
        const target = parseInt($this.text().replace(/[^\d]/g, ''));
        const duration = 2000;
        const step = target / (duration / 50);

        let current = 0;
        const timer = setInterval(function() {
            current += step;
            if (current >= target) {
                current = target;
                clearInterval(timer);
            }
            $this.text(Math.round(current) + ($this.text().includes('MB') ? 'MB' : ''));
        }, 50);
    });
}

// 上传页面初始化
function initUploadPage() {
    // 文件拖拽处理已在模板中实现
    console.log('Upload page initialized');
}

// 结果页面初始化
function initResultPage() {
    // 自动刷新处理已在模板中实现
    console.log('Result page initialized');
}

// 帮助页面初始化
function initHelpPage() {
    // 手风琴动画增强
    $('.accordion-button').on('click', function() {
        setTimeout(() => {
            const target = $($(this).attr('data-bs-target'));
            if (target.hasClass('show')) {
                $('html, body').animate({
                    scrollTop: target.offset().top - 100
                }, 300);
            }
        }, 150);
    });
}

// 工具函数

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化时间
function formatTime(date) {
    if (typeof date === 'string') {
        date = new Date(date);
    }
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// 显示提示消息
function showToast(message, type = 'info', duration = 3000) {
    const toastId = 'toast_' + Date.now();
    const iconMap = {
        'success': 'fas fa-check-circle',
        'error': 'fas fa-exclamation-circle',
        'warning': 'fas fa-exclamation-triangle',
        'info': 'fas fa-info-circle'
    };

    const colorMap = {
        'success': 'text-success',
        'error': 'text-danger',
        'warning': 'text-warning',
        'info': 'text-info'
    };

    const toastHtml = `
        <div id="${toastId}" class="toast align-items-center ${colorMap[type]} border-0" role="alert"
             style="position: fixed; top: 20px; right: 20px; z-index: 1051;">
            <div class="d-flex">
                <div class="toast-body">
                    <i class="${iconMap[type]} me-2"></i>
                    ${message}
                </div>
                <button type="button" class="btn-close me-2 m-auto" data-bs-dismiss="toast"></button>
            </div>
        </div>
    `;

    $('body').append(toastHtml);

    const toast = new bootstrap.Toast(document.getElementById(toastId), {
        delay: duration
    });

    toast.show();

    // 自动清理 DOM
    setTimeout(() => {
        $('#' + toastId).remove();
    }, duration + 1000);
}

// 确认对话框
function confirmAction(message, callback, title = '确认操作') {
    if (confirm(title + '\n\n' + message)) {
        callback();
    }
}

// 加载状态管理
function setLoadingState(element, loading = true, text = '处理中...') {
    const $el = $(element);

    if (loading) {
        $el.data('original-html', $el.html())
           .html(`<i class="fas fa-spinner fa-spin me-2"></i>${text}`)
           .prop('disabled', true)
           .addClass('disabled');
    } else {
        $el.html($el.data('original-html') || $el.html())
           .prop('disabled', false)
           .removeClass('disabled');
    }
}

// AJAX 错误处理
function handleAjaxError(xhr, textStatus, errorThrown) {
    let errorMessage = '请求失败';

    if (xhr.responseJSON && xhr.responseJSON.error) {
        errorMessage = xhr.responseJSON.error;
    } else if (xhr.responseText) {
        try {
            const response = JSON.parse(xhr.responseText);
            errorMessage = response.error || response.message || errorMessage;
        } catch (e) {
            errorMessage = xhr.responseText;
        }
    } else if (textStatus) {
        errorMessage = textStatus;
    }

    showToast(errorMessage, 'error');
}

// 设置全局 AJAX 错误处理
$(document).ajaxError(function(event, xhr, settings, thrownError) {
    // 只处理非自定义处理的错误
    if (!settings.skipGlobalError) {
        handleAjaxError(xhr, xhr.statusText, thrownError);
    }
});

// 进度条更新
function updateProgressBar(selector, progress, animated = true) {
    const $bar = $(selector);
    $bar.css('width', progress + '%')
        .attr('aria-valuenow', progress)
        .text(progress + '%');

    if (animated && !$bar.hasClass('progress-bar-animated')) {
        $bar.addClass('progress-bar-animated progress-bar-striped');
    } else if (!animated) {
        $bar.removeClass('progress-bar-animated progress-bar-striped');
    }
}

// 文件类型图标
function getFileTypeIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const iconMap = {
        'txt': 'fas fa-file-alt',
        'csv': 'fas fa-file-csv',
        'xlsx': 'fas fa-file-excel',
        'xls': 'fas fa-file-excel',
        'pdf': 'fas fa-file-pdf',
        'zip': 'fas fa-file-archive',
        'rar': 'fas fa-file-archive',
        '7z': 'fas fa-file-archive'
    };

    return iconMap[ext] || 'fas fa-file';
}

// 防抖函数
function debounce(func, wait, immediate) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            timeout = null;
            if (!immediate) func(...args);
        };
        const callNow = immediate && !timeout;
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
        if (callNow) func(...args);
    };
}

// 节流函数
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

// 页面可见性检测
function onVisibilityChange(callback) {
    document.addEventListener('visibilitychange', function() {
        callback(!document.hidden);
    });
}

// 浏览器功能检测
function checkBrowserFeatures() {
    const features = {
        fileAPI: !!(window.File && window.FileReader && window.FileList && window.Blob),
        dragAndDrop: 'draggable' in document.createElement('span'),
        localStorage: !!window.localStorage,
        sessionStorage: !!window.sessionStorage,
        webSocket: !!window.WebSocket
    };

    return features;
}

// 导出全局对象
window.WebBotUtils = {
    formatFileSize,
    formatTime,
    showToast,
    confirmAction,
    setLoadingState,
    handleAjaxError,
    updateProgressBar,
    getFileTypeIcon,
    debounce,
    throttle,
    onVisibilityChange,
    checkBrowserFeatures
};