/**
 * 主JavaScript文件
 * 包含导航菜单和交互功能
 */

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', function() {
    // 处理移动菜单
    setupMobileMenu();
    
    // 添加特效到功能卡片
    setupFeatureCards();
    
    // 添加平滑滚动
    setupSmoothScrolling();
    
    // 添加代码高亮
    highlightCodeBlocks();
});

/**
 * 设置移动菜单功能
 */
function setupMobileMenu() {
    const menuToggle = document.querySelector('.menu-toggle');
    const navbarNav = document.querySelector('.navbar-nav');
    
    if (menuToggle && navbarNav) {
        menuToggle.addEventListener('click', function() {
            // 切换导航菜单的显示状态
            navbarNav.classList.toggle('active');
            
            // 切换汉堡按钮的动画状态
            const spans = menuToggle.querySelectorAll('span');
            spans.forEach(span => span.classList.toggle('active'));
        });
        
        // 点击菜单项后关闭菜单
        const navItems = navbarNav.querySelectorAll('a');
        navItems.forEach(item => {
            item.addEventListener('click', function() {
                if (window.innerWidth <= 768) {
                    navbarNav.classList.remove('active');
                    
                    const spans = menuToggle.querySelectorAll('span');
                    spans.forEach(span => span.classList.remove('active'));
                }
            });
        });
    }
}

/**
 * 设置功能卡片特效
 */
function setupFeatureCards() {
    const featureCards = document.querySelectorAll('.feature-card');
    
    featureCards.forEach((card, index) => {
        // 设置动画顺序
        card.style.setProperty('--animation-order', index);
        
        // 添加鼠标悬停效果
        card.addEventListener('mouseenter', function() {
            this.style.transform = 'translateY(-10px)';
        });
        
        card.addEventListener('mouseleave', function() {
            this.style.transform = '';
        });
    });
}

/**
 * 设置平滑滚动
 */
function setupSmoothScrolling() {
    // 获取所有页内链接
    const internalLinks = document.querySelectorAll('a[href^="#"]');
    
    internalLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            const targetId = this.getAttribute('href');
            
            // 确保目标存在
            if (targetId !== '#') {
                const targetElement = document.querySelector(targetId);
                
                if (targetElement) {
                    e.preventDefault();
                    
                    // 平滑滚动到目标位置
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            }
        });
    });
}

/**
 * 转义正则表达式特殊字符
 */
function escapeRegExp(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

/**
 * 转义HTML
 */
function escapeHTML(text) {
    return text.replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

/**
 * 为代码块添加高亮效果
 */
function highlightCodeBlocks() {
    const codeBlocks = document.querySelectorAll('.code-block');
    
    codeBlocks.forEach(block => {
        // 避免重复处理
        if (block.dataset.processed) {
            return;
        }
        
        // 获取代码语言
        const language = block.dataset.lang || 'text';
        
        // 获取代码内容
        const code = block.textContent || block.innerText;
        
        try {
            // 解码HTML实体
            let processedCode = code.replace(/&lt;/g, '<')
                                   .replace(/&gt;/g, '>')
                                   .replace(/&amp;/g, '&')
                                   .replace(/&quot;/g, '"')
                                   .replace(/&#039;/g, "'")
                                   .replace(/&#39;/g, "'");
            
            // 预处理：保存模板标签
            const templateTags = [];
            let templateCount = 0;
            processedCode = processedCode.replace(/{{[\s\S]*?}}/g, match => {
                const placeholder = `__TEMPLATE_${templateCount++}__`;
                templateTags.push({ placeholder, content: match });
                return placeholder;
            });
            
            // 使用简单的词法分析器处理代码
            const tokens = simpleTokenize(processedCode);
            
            // 将分词结果转换为HTML
            let highlightedCode = '';
            for (const token of tokens) {
                const escapedValue = escapeHTML(token.value);
                
                // 检查是否是模板标签占位符
                const templateTag = templateTags.find(t => t.placeholder === token.value);
                if (templateTag) {
                    highlightedCode += `<span class="code-template">${escapeHTML(templateTag.content)}</span>`;
                    continue;
                }
                
                switch (token.type) {
                    case 'keyword':
                        highlightedCode += `<span class="code-keyword">${escapedValue}</span>`;
                        break;
                    case 'type':
                        highlightedCode += `<span class="code-type">${escapedValue}</span>`;
                        break;
                    case 'stdlib':
                        highlightedCode += `<span class="code-stdlib">${escapedValue}</span>`;
                        break;
                    case 'string':
                        highlightedCode += `<span class="code-string">${escapedValue}</span>`;
                        break;
                    case 'comment':
                        highlightedCode += `<span class="code-comment">${escapedValue}</span>`;
                        break;
                    case 'number':
                        highlightedCode += `<span class="code-number">${escapedValue}</span>`;
                        break;
                    case 'function':
                        highlightedCode += `<span class="code-function">${escapedValue}</span>`;
                        break;
                    case 'punctuation':
                        highlightedCode += `<span class="code-punctuation">${escapedValue}</span>`;
                        break;
                    case 'interface':
                        highlightedCode += `<span class="code-interface">${escapedValue}</span>`;
                        break;
                    case 'variable':
                        highlightedCode += `<span class="code-variable">${escapedValue}</span>`;
                        break;
                    default:
                        highlightedCode += escapedValue;
                }
            }
            
            // 添加行号
            const lines = highlightedCode.split('\n');
            const numberedLines = lines.map((line, i) => {
                const lineNum = i + 1;
                return `<span class="code-line"><span class="code-line-number">${lineNum}</span><span class="code-line-content">${line || ' '}</span></span>`;
            }).join('\n');
            
            // 添加语言标识
            const langLabel = language !== 'text' ? 
                `<div class="code-lang-label">${language.toUpperCase()}</div>` : '';
            
            // 更新代码块内容
            block.innerHTML = `${langLabel}<pre class="code-highlighted">${numberedLines}</pre>`;
            block.dataset.processed = 'true';
            block.classList.add(`lang-${language}`);
        } catch (error) {
            console.error('代码高亮处理错误:', error);
            // 出错时使用原始内容，确保显示
            const escapedCode = escapeHTML(code);
            block.innerHTML = `<pre>${escapedCode}</pre>`;
            block.dataset.processed = 'true';
        }
    });
}

/**
 * 简单的词法分析器，避免嵌套标签问题
 */
function simpleTokenize(code) {
    const tokens = [];
    let pos = 0;
    
    // 关键字和特殊标识符列表
    const keywords = [
        'break', 'default', 'func', 'interface', 'select', 'case', 'defer', 'go', 
        'map', 'struct', 'chan', 'else', 'goto', 'package', 'switch', 'const', 
        'fallthrough', 'if', 'range', 'type', 'continue', 'for', 'import', 
        'return', 'var', 'In', 'GET', 'POST', 'PUT', 'DELETE'
    ];
    
    // 标准库列表
    const stdlibs = [
        'fmt', 'http', 'gin', 'io', 'os', 'time', 'strings', 'strconv', 'math', 
        'context', 'errors', 'fx', 'sync', 'log', 'database/sql', 'response', 
        'middleware', 'config', 'template'
    ];
    
    // 特殊类型
    const specialTypes = ['string', 'int', 'bool', 'map', 'interface', 'Context'];
    
    // 正则表达式
    const regexes = {
        whitespace: /^\s+/,
        comment: /^\/\/.*$/,
        string: /^"(?:\\.|[^"\\])*"/,
        number: /^\d+(\.\d+)?/,
        identifier: /^[a-zA-Z_]\w*/,
        punctuation: /^[{}[\](),.;:=<>+\-*/%&|^!~?]/
    };
    
    // 处理每个字符，直到结束
    while (pos < code.length) {
        let matched = false;
        
        // 匹配空白字符
        const wsMatch = code.slice(pos).match(regexes.whitespace);
        if (wsMatch) {
            tokens.push({ type: 'whitespace', value: wsMatch[0] });
            pos += wsMatch[0].length;
            matched = true;
            continue;
        }
        
        // 匹配注释
        const commentMatch = code.slice(pos).match(regexes.comment);
        if (commentMatch) {
            tokens.push({ type: 'comment', value: commentMatch[0] });
            pos += commentMatch[0].length;
            matched = true;
            continue;
        }
        
        // 匹配字符串
        const stringMatch = code.slice(pos).match(regexes.string);
        if (stringMatch) {
            tokens.push({ type: 'string', value: stringMatch[0] });
            pos += stringMatch[0].length;
            matched = true;
            continue;
        }
        
        // 匹配数字
        const numberMatch = code.slice(pos).match(regexes.number);
        if (numberMatch) {
            tokens.push({ type: 'number', value: numberMatch[0] });
            pos += numberMatch[0].length;
            matched = true;
            continue;
        }
        
        // 匹配标识符
        const idMatch = code.slice(pos).match(regexes.identifier);
        if (idMatch) {
            const value = idMatch[0];
            
            // 检查是否是关键字
            if (keywords.includes(value)) {
                tokens.push({ type: 'keyword', value });
            }
            // 检查是否是标准库
            else if (stdlibs.includes(value)) {
                tokens.push({ type: 'stdlib', value });
            }
            // 检查是否是特殊类型
            else if (specialTypes.includes(value)) {
                tokens.push({ type: 'type', value });
            }
            // 检查是否是类型名（大写开头）
            else if (/^[A-Z]/.test(value)) {
                tokens.push({ type: 'type', value });
            }
            // 检查是否是函数名（后面跟括号）
            else if (pos + value.length < code.length && code[pos + value.length] === '(') {
                tokens.push({ type: 'function', value });
            }
            // 其他标识符
            else {
                tokens.push({ type: 'identifier', value });
            }
            
            pos += value.length;
            matched = true;
            continue;
        }
        
        // 匹配标点符号
        const punctMatch = code.slice(pos).match(regexes.punctuation);
        if (punctMatch) {
            tokens.push({ type: 'punctuation', value: punctMatch[0] });
            pos += punctMatch[0].length;
            matched = true;
            continue;
        }
        
        // 如果没有匹配到任何规则，处理单个字符
        if (!matched) {
            tokens.push({ type: 'text', value: code[pos] });
            pos++;
        }
    }
    
    return tokens;
} 