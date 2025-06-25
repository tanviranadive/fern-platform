# GitHub Pages Setup Guide

## ✅ **Files Ready for GitHub Pages**

I've created `index.html` with all the correct paths for GitHub Pages deployment.

## 🚀 **Deployment Steps**

### **1. Enable GitHub Pages**
1. Go to your repository: `https://github.com/guidewire-oss/fern-platform`
2. Click **Settings** tab
3. Scroll down to **Pages** section
4. Under **Source**, select:
   - **Deploy from a branch**
   - Branch: **main**
   - Folder: **/ (root)**
5. Click **Save**

### **2. Wait for Deployment**
- GitHub will show a notification: "Your site is ready to be published"
- Initial deployment takes 5-10 minutes
- You'll get a green checkmark when ready

### **3. Access Your Landing Page**
Your landing page will be available at:
```
https://guidewire-oss.github.io/fern-platform/
```

## 📁 **File Structure**

```
fern-platform/
├── index.html          ← Landing page (GitHub Pages homepage)
├── landing.html        ← Original landing page (backup)
├── README.md           ← Developer documentation
├── docs/
│   └── images/         ← All screenshots and assets
└── ...
```

## 🎯 **Result**

**Two Different URLs, Two Different Audiences:**

✅ **Marketing Landing Page**: `https://guidewire-oss.github.io/fern-platform/`
- Beautiful, visual landing page
- Perfect for sharing on social media
- Drives user engagement and adoption

✅ **Developer Documentation**: `https://github.com/guidewire-oss/fern-platform`
- Technical README with setup instructions
- Perfect for developers and contributors
- Shows repository structure and code

## 🔧 **Image Paths Fixed**

All images now use GitHub Pages-compatible paths:
- **Before**: `docs/images/logo-color.png`
- **After**: `/fern-platform/docs/images/logo-color.png`

Social media sharing (Open Graph/Twitter) uses full URLs:
- `https://guidewire-oss.github.io/fern-platform/docs/images/test-uber-treemap.png`

## 📊 **Analytics Setup**

Once your page is live, you can:
1. Add Google Analytics tracking ID to `index.html`
2. Monitor traffic from the landing page vs repository
3. Track conversion from landing page to GitHub repo

## 🚀 **Next Steps**

1. **Commit and push** the `index.html` file
2. **Enable GitHub Pages** in repository settings
3. **Share the landing page URL** for marketing
4. **Keep README.md** for developers
5. **Monitor analytics** to see engagement

## 🎉 **Benefits**

- **Professional Landing Page**: First-class marketing presence
- **SEO Optimized**: Better search engine discovery
- **Social Media Ready**: Rich previews when shared
- **Developer Friendly**: README still serves technical audience
- **Free Hosting**: No need for external hosting services

Your landing page is now ready for prime time! 🌟