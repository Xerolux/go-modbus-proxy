import{c as w,o as m,a as V,x as d,B as H,ak as Z,aE as R,a8 as U,a9 as O,v as z,h as K,y as f,d as B,i as J,n as j,p as D,z as k,aF as M}from"./index-DorqM7ya.js";import{s as Q}from"./index-BX33RUWu.js";import{a as X,f as Y}from"./index-ChQ6boSU.js";import{a as ee}from"./index-DMRAzI-C.js";import{s as ne}from"./index-DEJ4SokQ.js";import{s as te}from"./index-CII8oppT.js";var W={name:"AngleUpIcon",extends:X};function ie(n){return ae(n)||se(n)||ue(n)||re()}function re(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function ue(n,e){if(n){if(typeof n=="string")return A(n,e);var t={}.toString.call(n).slice(8,-1);return t==="Object"&&n.constructor&&(t=n.constructor.name),t==="Map"||t==="Set"?Array.from(n):t==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t)?A(n,e):void 0}}function se(n){if(typeof Symbol<"u"&&n[Symbol.iterator]!=null||n["@@iterator"]!=null)return Array.from(n)}function ae(n){if(Array.isArray(n))return A(n)}function A(n,e){(e==null||e>n.length)&&(e=n.length);for(var t=0,i=Array(e);t<e;t++)i[t]=n[t];return i}function oe(n,e,t,i,u,r){return m(),w("svg",d({width:"14",height:"14",viewBox:"0 0 14 14",fill:"none",xmlns:"http://www.w3.org/2000/svg"},n.pti()),ie(e[0]||(e[0]=[V("path",{d:"M10.4134 9.49931C10.3148 9.49977 10.2172 9.48055 10.1262 9.44278C10.0352 9.405 9.95263 9.34942 9.88338 9.27931L6.88338 6.27931L3.88338 9.27931C3.73811 9.34946 3.57409 9.3709 3.41567 9.34044C3.25724 9.30999 3.11286 9.22926 3.00395 9.11025C2.89504 8.99124 2.82741 8.84028 2.8111 8.67978C2.79478 8.51928 2.83065 8.35781 2.91338 8.21931L6.41338 4.71931C6.55401 4.57886 6.74463 4.49997 6.94338 4.49997C7.14213 4.49997 7.33276 4.57886 7.47338 4.71931L10.9734 8.21931C11.1138 8.35994 11.1927 8.55056 11.1927 8.74931C11.1927 8.94806 11.1138 9.13868 10.9734 9.27931C10.9007 9.35315 10.8132 9.41089 10.7168 9.44879C10.6203 9.48669 10.5169 9.5039 10.4134 9.49931Z",fill:"currentColor"},null,-1)])),16)}W.render=oe;var le=`
    .p-inputnumber {
        display: inline-flex;
        position: relative;
    }

    .p-inputnumber-button {
        display: flex;
        align-items: center;
        justify-content: center;
        flex: 0 0 auto;
        cursor: pointer;
        background: dt('inputnumber.button.background');
        color: dt('inputnumber.button.color');
        width: dt('inputnumber.button.width');
        transition:
            background dt('inputnumber.transition.duration'),
            color dt('inputnumber.transition.duration'),
            border-color dt('inputnumber.transition.duration'),
            outline-color dt('inputnumber.transition.duration');
    }

    .p-inputnumber-button:disabled {
        cursor: auto;
    }

    .p-inputnumber-button:not(:disabled):hover {
        background: dt('inputnumber.button.hover.background');
        color: dt('inputnumber.button.hover.color');
    }

    .p-inputnumber-button:not(:disabled):active {
        background: dt('inputnumber.button.active.background');
        color: dt('inputnumber.button.active.color');
    }

    .p-inputnumber-stacked .p-inputnumber-button {
        position: relative;
        flex: 1 1 auto;
        border: 0 none;
    }

    .p-inputnumber-stacked .p-inputnumber-button-group {
        display: flex;
        flex-direction: column;
        position: absolute;
        inset-block-start: 1px;
        inset-inline-end: 1px;
        height: calc(100% - 2px);
        z-index: 1;
    }

    .p-inputnumber-stacked .p-inputnumber-increment-button {
        padding: 0;
        border-start-end-radius: calc(dt('inputnumber.button.border.radius') - 1px);
    }

    .p-inputnumber-stacked .p-inputnumber-decrement-button {
        padding: 0;
        border-end-end-radius: calc(dt('inputnumber.button.border.radius') - 1px);
    }

    .p-inputnumber-stacked .p-inputnumber-input {
        padding-inline-end: calc(dt('inputnumber.button.width') + dt('form.field.padding.x'));
    }

    .p-inputnumber-horizontal .p-inputnumber-button {
        border: 1px solid dt('inputnumber.button.border.color');
    }

    .p-inputnumber-horizontal .p-inputnumber-button:hover {
        border-color: dt('inputnumber.button.hover.border.color');
    }

    .p-inputnumber-horizontal .p-inputnumber-button:active {
        border-color: dt('inputnumber.button.active.border.color');
    }

    .p-inputnumber-horizontal .p-inputnumber-increment-button {
        order: 3;
        border-start-end-radius: dt('inputnumber.button.border.radius');
        border-end-end-radius: dt('inputnumber.button.border.radius');
        border-inline-start: 0 none;
    }

    .p-inputnumber-horizontal .p-inputnumber-input {
        order: 2;
        border-radius: 0;
    }

    .p-inputnumber-horizontal .p-inputnumber-decrement-button {
        order: 1;
        border-start-start-radius: dt('inputnumber.button.border.radius');
        border-end-start-radius: dt('inputnumber.button.border.radius');
        border-inline-end: 0 none;
    }

    .p-floatlabel:has(.p-inputnumber-horizontal) label {
        margin-inline-start: dt('inputnumber.button.width');
    }

    .p-inputnumber-vertical {
        flex-direction: column;
    }

    .p-inputnumber-vertical .p-inputnumber-button {
        border: 1px solid dt('inputnumber.button.border.color');
        padding: dt('inputnumber.button.vertical.padding');
    }

    .p-inputnumber-vertical .p-inputnumber-button:hover {
        border-color: dt('inputnumber.button.hover.border.color');
    }

    .p-inputnumber-vertical .p-inputnumber-button:active {
        border-color: dt('inputnumber.button.active.border.color');
    }

    .p-inputnumber-vertical .p-inputnumber-increment-button {
        order: 1;
        border-start-start-radius: dt('inputnumber.button.border.radius');
        border-start-end-radius: dt('inputnumber.button.border.radius');
        width: 100%;
        border-block-end: 0 none;
    }

    .p-inputnumber-vertical .p-inputnumber-input {
        order: 2;
        border-radius: 0;
        text-align: center;
    }

    .p-inputnumber-vertical .p-inputnumber-decrement-button {
        order: 3;
        border-end-start-radius: dt('inputnumber.button.border.radius');
        border-end-end-radius: dt('inputnumber.button.border.radius');
        width: 100%;
        border-block-start: 0 none;
    }

    .p-inputnumber-input {
        flex: 1 1 auto;
    }

    .p-inputnumber-fluid {
        width: 100%;
    }

    .p-inputnumber-fluid .p-inputnumber-input {
        width: 1%;
    }

    .p-inputnumber-fluid.p-inputnumber-vertical .p-inputnumber-input {
        width: 100%;
    }

    .p-inputnumber:has(.p-inputtext-sm) .p-inputnumber-button .p-icon {
        font-size: dt('form.field.sm.font.size');
        width: dt('form.field.sm.font.size');
        height: dt('form.field.sm.font.size');
    }

    .p-inputnumber:has(.p-inputtext-lg) .p-inputnumber-button .p-icon {
        font-size: dt('form.field.lg.font.size');
        width: dt('form.field.lg.font.size');
        height: dt('form.field.lg.font.size');
    }

    .p-inputnumber-clear-icon {
        position: absolute;
        top: 50%;
        margin-top: -0.5rem;
        cursor: pointer;
        inset-inline-end: dt('form.field.padding.x');
        color: dt('form.field.icon.color');
    }

    .p-inputnumber:has(.p-inputnumber-clear-icon) .p-inputnumber-input {
        padding-inline-end: calc((dt('form.field.padding.x') * 2) + dt('icon.size'));
    }

    .p-inputnumber-stacked .p-inputnumber-clear-icon {
        inset-inline-end: calc(dt('inputnumber.button.width') + dt('form.field.padding.x'));
    }

    .p-inputnumber-stacked:has(.p-inputnumber-clear-icon) .p-inputnumber-input {
        padding-inline-end: calc(dt('inputnumber.button.width') + (dt('form.field.padding.x') * 2) + dt('icon.size'));
    }

    .p-inputnumber-horizontal .p-inputnumber-clear-icon {
        inset-inline-end: calc(dt('inputnumber.button.width') + dt('form.field.padding.x'));
    }
`,ce={root:function(e){var t=e.instance,i=e.props;return["p-inputnumber p-component p-inputwrapper",{"p-invalid":t.$invalid,"p-inputwrapper-filled":t.$filled||i.allowEmpty===!1,"p-inputwrapper-focus":t.focused,"p-inputnumber-stacked":i.showButtons&&i.buttonLayout==="stacked","p-inputnumber-horizontal":i.showButtons&&i.buttonLayout==="horizontal","p-inputnumber-vertical":i.showButtons&&i.buttonLayout==="vertical","p-inputnumber-fluid":t.$fluid}]},pcInputText:"p-inputnumber-input",clearIcon:"p-inputnumber-clear-icon",buttonGroup:"p-inputnumber-button-group",incrementButton:function(e){var t=e.instance,i=e.props;return["p-inputnumber-button p-inputnumber-increment-button",{"p-disabled":i.showButtons&&i.max!==null&&t.maxBoundry()}]},decrementButton:function(e){var t=e.instance,i=e.props;return["p-inputnumber-button p-inputnumber-decrement-button",{"p-disabled":i.showButtons&&i.min!==null&&t.minBoundry()}]}},pe=H.extend({name:"inputnumber",style:le,classes:ce}),de={name:"BaseInputNumber",extends:te,props:{format:{type:Boolean,default:!0},showButtons:{type:Boolean,default:!1},buttonLayout:{type:String,default:"stacked"},incrementButtonClass:{type:String,default:null},decrementButtonClass:{type:String,default:null},incrementButtonIcon:{type:String,default:void 0},incrementIcon:{type:String,default:void 0},decrementButtonIcon:{type:String,default:void 0},decrementIcon:{type:String,default:void 0},locale:{type:String,default:void 0},localeMatcher:{type:String,default:void 0},mode:{type:String,default:"decimal"},prefix:{type:String,default:null},suffix:{type:String,default:null},currency:{type:String,default:void 0},currencyDisplay:{type:String,default:void 0},useGrouping:{type:Boolean,default:!0},minFractionDigits:{type:Number,default:void 0},maxFractionDigits:{type:Number,default:void 0},roundingMode:{type:String,default:"halfExpand",validator:function(e){return["ceil","floor","expand","trunc","halfCeil","halfFloor","halfExpand","halfTrunc","halfEven"].includes(e)}},min:{type:Number,default:null},max:{type:Number,default:null},step:{type:Number,default:1},allowEmpty:{type:Boolean,default:!0},highlightOnFocus:{type:Boolean,default:!1},showClear:{type:Boolean,default:!1},readonly:{type:Boolean,default:!1},placeholder:{type:String,default:null},inputId:{type:String,default:null},inputClass:{type:[String,Object],default:null},inputStyle:{type:Object,default:null},ariaLabelledby:{type:String,default:null},ariaLabel:{type:String,default:null},required:{type:Boolean,default:!1}},style:pe,provide:function(){return{$pcInputNumber:this,$parentInstance:this}}};function C(n){"@babel/helpers - typeof";return C=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(e){return typeof e}:function(e){return e&&typeof Symbol=="function"&&e.constructor===Symbol&&e!==Symbol.prototype?"symbol":typeof e},C(n)}function G(n,e){var t=Object.keys(n);if(Object.getOwnPropertySymbols){var i=Object.getOwnPropertySymbols(n);e&&(i=i.filter(function(u){return Object.getOwnPropertyDescriptor(n,u).enumerable})),t.push.apply(t,i)}return t}function _(n){for(var e=1;e<arguments.length;e++){var t=arguments[e]!=null?arguments[e]:{};e%2?G(Object(t),!0).forEach(function(i){F(n,i,t[i])}):Object.getOwnPropertyDescriptors?Object.defineProperties(n,Object.getOwnPropertyDescriptors(t)):G(Object(t)).forEach(function(i){Object.defineProperty(n,i,Object.getOwnPropertyDescriptor(t,i))})}return n}function F(n,e,t){return(e=he(e))in n?Object.defineProperty(n,e,{value:t,enumerable:!0,configurable:!0,writable:!0}):n[e]=t,n}function he(n){var e=fe(n,"string");return C(e)=="symbol"?e:e+""}function fe(n,e){if(C(n)!="object"||!n)return n;var t=n[Symbol.toPrimitive];if(t!==void 0){var i=t.call(n,e);if(C(i)!="object")return i;throw new TypeError("@@toPrimitive must return a primitive value.")}return(e==="string"?String:Number)(n)}function me(n){return ve(n)||ye(n)||ge(n)||be()}function be(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function ge(n,e){if(n){if(typeof n=="string")return E(n,e);var t={}.toString.call(n).slice(8,-1);return t==="Object"&&n.constructor&&(t=n.constructor.name),t==="Map"||t==="Set"?Array.from(n):t==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t)?E(n,e):void 0}}function ye(n){if(typeof Symbol<"u"&&n[Symbol.iterator]!=null||n["@@iterator"]!=null)return Array.from(n)}function ve(n){if(Array.isArray(n))return E(n)}function E(n,e){(e==null||e>n.length)&&(e=n.length);for(var t=0,i=Array(e);t<e;t++)i[t]=n[t];return i}var Ie={name:"InputNumber",extends:de,inheritAttrs:!1,emits:["input","focus","blur"],inject:{$pcFluid:{default:null}},numberFormat:null,_numeral:null,_decimal:null,_group:null,_minusSign:null,_currency:null,_suffix:null,_prefix:null,_index:null,groupChar:"",isSpecialChar:null,prefixChar:null,suffixChar:null,timer:null,data:function(){return{d_modelValue:this.d_value,focused:!1}},watch:{d_value:{immediate:!0,handler:function(e){var t;this.d_modelValue=e,(t=this.$refs.clearIcon)!==null&&t!==void 0&&(t=t.$el)!==null&&t!==void 0&&t.style&&(this.$refs.clearIcon.$el.style.display=U(e)?"none":"block")}},locale:function(e,t){this.updateConstructParser(e,t)},localeMatcher:function(e,t){this.updateConstructParser(e,t)},mode:function(e,t){this.updateConstructParser(e,t)},currency:function(e,t){this.updateConstructParser(e,t)},currencyDisplay:function(e,t){this.updateConstructParser(e,t)},useGrouping:function(e,t){this.updateConstructParser(e,t)},minFractionDigits:function(e,t){this.updateConstructParser(e,t)},maxFractionDigits:function(e,t){this.updateConstructParser(e,t)},suffix:function(e,t){this.updateConstructParser(e,t)},prefix:function(e,t){this.updateConstructParser(e,t)}},created:function(){this.constructParser()},mounted:function(){var e;(e=this.$refs.clearIcon)!==null&&e!==void 0&&(e=e.$el)!==null&&e!==void 0&&e.style&&(this.$refs.clearIcon.$el.style.display=this.$filled?"block":"none")},methods:{getOptions:function(){return{localeMatcher:this.localeMatcher,style:this.mode,currency:this.currency,currencyDisplay:this.currencyDisplay,useGrouping:this.useGrouping,minimumFractionDigits:this.minFractionDigits,maximumFractionDigits:this.maxFractionDigits,roundingMode:this.roundingMode}},constructParser:function(){this.numberFormat=new Intl.NumberFormat(this.locale,this.getOptions());var e=me(new Intl.NumberFormat(this.locale,{useGrouping:!1}).format(9876543210)).reverse(),t=new Map(e.map(function(i,u){return[i,u]}));this._numeral=new RegExp("[".concat(e.join(""),"]"),"g"),this._group=this.getGroupingExpression(),this._minusSign=this.getMinusSignExpression(),this._currency=this.getCurrencyExpression(),this._decimal=this.getDecimalExpression(),this._suffix=this.getSuffixExpression(),this._prefix=this.getPrefixExpression(),this._index=function(i){return t.get(i)}},updateConstructParser:function(e,t){e!==t&&this.constructParser()},escapeRegExp:function(e){return e.replace(/[-[\]{}()*+?.,\\^$|#\s]/g,"\\$&")},getDecimalExpression:function(){var e=new Intl.NumberFormat(this.locale,_(_({},this.getOptions()),{},{useGrouping:!1}));return new RegExp("[".concat(e.format(1.1).replace(this._currency,"").trim().replace(this._numeral,""),"]"),"g")},getGroupingExpression:function(){var e=new Intl.NumberFormat(this.locale,{useGrouping:!0});return this.groupChar=e.format(1e6).trim().replace(this._numeral,"").charAt(0),new RegExp("[".concat(this.groupChar,"]"),"g")},getMinusSignExpression:function(){var e=new Intl.NumberFormat(this.locale,{useGrouping:!1});return new RegExp("[".concat(e.format(-1).trim().replace(this._numeral,""),"]"),"g")},getCurrencyExpression:function(){if(this.currency){var e=new Intl.NumberFormat(this.locale,{style:"currency",currency:this.currency,currencyDisplay:this.currencyDisplay,minimumFractionDigits:0,maximumFractionDigits:0,roundingMode:this.roundingMode});return new RegExp("[".concat(e.format(1).replace(/\s/g,"").replace(this._numeral,"").replace(this._group,""),"]"),"g")}return new RegExp("[]","g")},getPrefixExpression:function(){if(this.prefix)this.prefixChar=this.prefix;else{var e=new Intl.NumberFormat(this.locale,{style:this.mode,currency:this.currency,currencyDisplay:this.currencyDisplay});this.prefixChar=e.format(1).split("1")[0]}return new RegExp("".concat(this.escapeRegExp(this.prefixChar||"")),"g")},getSuffixExpression:function(){if(this.suffix)this.suffixChar=this.suffix;else{var e=new Intl.NumberFormat(this.locale,{style:this.mode,currency:this.currency,currencyDisplay:this.currencyDisplay,minimumFractionDigits:0,maximumFractionDigits:0,roundingMode:this.roundingMode});this.suffixChar=e.format(1).split("1")[1]}return new RegExp("".concat(this.escapeRegExp(this.suffixChar||"")),"g")},formatValue:function(e){if(e!=null){if(e==="-")return e;if(this.format){var t=new Intl.NumberFormat(this.locale,this.getOptions()),i=t.format(e);return this.prefix&&(i=this.prefix+i),this.suffix&&(i=i+this.suffix),i}return e.toString()}return""},parseValue:function(e){var t=e.replace(this._suffix,"").replace(this._prefix,"").trim().replace(/\s/g,"").replace(this._currency,"").replace(this._group,"").replace(this._minusSign,"-").replace(this._decimal,".").replace(this._numeral,this._index);if(t){if(t==="-")return t;var i=+t;return isNaN(i)?null:i}return null},repeat:function(e,t,i){var u=this;if(!this.readonly){var r=t||500;this.clearTimer(),this.timer=setTimeout(function(){u.repeat(e,40,i)},r),this.spin(e,i)}},addWithPrecision:function(e,t){var i=e.toString(),u=t.toString(),r=i.includes(".")?i.split(".")[1].length:0,s=u.includes(".")?u.split(".")[1].length:0,a=Math.max(r,s),c=Math.pow(10,a);return Math.round((e+t)*c)/c},spin:function(e,t){if(this.$refs.input){var i=this.step*t,u=this.parseValue(this.$refs.input.$el.value)||0,r=this.validateValue(this.addWithPrecision(u,i));this.updateInput(r,null,"spin"),this.updateModel(e,r),this.handleOnInput(e,u,r)}},onUpButtonMouseDown:function(e){this.disabled||(this.$refs.input.$el.focus(),this.repeat(e,null,1),e.preventDefault())},onUpButtonMouseUp:function(){this.disabled||this.clearTimer()},onUpButtonMouseLeave:function(){this.disabled||this.clearTimer()},onUpButtonKeyUp:function(){this.disabled||this.clearTimer()},onUpButtonKeyDown:function(e){(e.code==="Space"||e.code==="Enter"||e.code==="NumpadEnter")&&this.repeat(e,null,1)},onDownButtonMouseDown:function(e){this.disabled||(this.$refs.input.$el.focus(),this.repeat(e,null,-1),e.preventDefault())},onDownButtonMouseUp:function(){this.disabled||this.clearTimer()},onDownButtonMouseLeave:function(){this.disabled||this.clearTimer()},onDownButtonKeyUp:function(){this.disabled||this.clearTimer()},onDownButtonKeyDown:function(e){(e.code==="Space"||e.code==="Enter"||e.code==="NumpadEnter")&&this.repeat(e,null,-1)},onUserInput:function(){this.isSpecialChar&&(this.$refs.input.$el.value=this.lastValue),this.isSpecialChar=!1},onInputKeyDown:function(e){if(!this.readonly&&!e.isComposing){if(e.altKey||e.ctrlKey||e.metaKey){this.isSpecialChar=!0,this.lastValue=this.$refs.input.$el.value;return}this.lastValue=e.target.value;var t=e.target.selectionStart,i=e.target.selectionEnd,u=i-t,r=e.target.value,s=null,a=e.code||e.key;switch(a){case"ArrowUp":this.spin(e,1),e.preventDefault();break;case"ArrowDown":this.spin(e,-1),e.preventDefault();break;case"ArrowLeft":if(u>1){var c=this.isNumeralChar(r.charAt(t))?t+1:t+2;this.$refs.input.$el.setSelectionRange(c,c)}else this.isNumeralChar(r.charAt(t-1))||e.preventDefault();break;case"ArrowRight":if(u>1){var p=i-1;this.$refs.input.$el.setSelectionRange(p,p)}else this.isNumeralChar(r.charAt(t))||e.preventDefault();break;case"Tab":case"Enter":case"NumpadEnter":s=this.validateValue(this.parseValue(r)),this.$refs.input.$el.value=this.formatValue(s),this.$refs.input.$el.setAttribute("aria-valuenow",s),this.updateModel(e,s);break;case"Backspace":{if(e.preventDefault(),t===i){t>=r.length&&this.suffixChar!==null&&(t=r.length-this.suffixChar.length,this.$refs.input.$el.setSelectionRange(t,t));var b=r.charAt(t-1),y=this.getDecimalCharIndexes(r),o=y.decimalCharIndex,l=y.decimalCharIndexWithoutPrefix;if(this.isNumeralChar(b)){var v=this.getDecimalLength(r);if(this._group.test(b))this._group.lastIndex=0,s=r.slice(0,t-2)+r.slice(t-1);else if(this._decimal.test(b))this._decimal.lastIndex=0,v?this.$refs.input.$el.setSelectionRange(t-1,t-1):s=r.slice(0,t-1)+r.slice(t);else if(o>0&&t>o){var I=this.isDecimalMode()&&(this.minFractionDigits||0)<v?"":"0";s=r.slice(0,t-1)+I+r.slice(t)}else l===1?(s=r.slice(0,t-1)+"0"+r.slice(t),s=this.parseValue(s)>0?s:""):s=r.slice(0,t-1)+r.slice(t)}this.updateValue(e,s,null,"delete-single")}else s=this.deleteRange(r,t,i),this.updateValue(e,s,null,"delete-range");break}case"Delete":if(e.preventDefault(),t===i){var g=r.charAt(t),h=this.getDecimalCharIndexes(r),x=h.decimalCharIndex,$=h.decimalCharIndexWithoutPrefix;if(this.isNumeralChar(g)){var S=this.getDecimalLength(r);if(this._group.test(g))this._group.lastIndex=0,s=r.slice(0,t)+r.slice(t+2);else if(this._decimal.test(g))this._decimal.lastIndex=0,S?this.$refs.input.$el.setSelectionRange(t+1,t+1):s=r.slice(0,t)+r.slice(t+1);else if(x>0&&t>x){var P=this.isDecimalMode()&&(this.minFractionDigits||0)<S?"":"0";s=r.slice(0,t)+P+r.slice(t+1)}else $===1?(s=r.slice(0,t)+"0"+r.slice(t+1),s=this.parseValue(s)>0?s:""):s=r.slice(0,t)+r.slice(t+1)}this.updateValue(e,s,null,"delete-back-single")}else s=this.deleteRange(r,t,i),this.updateValue(e,s,null,"delete-range");break;case"Home":e.preventDefault(),O(this.min)&&this.updateModel(e,this.min);break;case"End":e.preventDefault(),O(this.max)&&this.updateModel(e,this.max);break}}},onInputKeyPress:function(e){if(!this.readonly){var t=e.key,i=this.isDecimalSign(t),u=this.isMinusSign(t);e.code!=="Enter"&&e.preventDefault(),(Number(t)>=0&&Number(t)<=9||u||i)&&this.insert(e,t,{isDecimalSign:i,isMinusSign:u})}},onPaste:function(e){if(!this.readonly){e.preventDefault();var t=(e.clipboardData||window.clipboardData).getData("Text");if(!(this.inputId==="integeronly"&&/[^\d-]/.test(t))&&t){var i=this.parseValue(t);i!=null&&this.insert(e,i.toString())}}},onClearClick:function(e){this.updateModel(e,null),this.$refs.input.$el.focus()},allowMinusSign:function(){return this.min===null||this.min<0},isMinusSign:function(e){return this._minusSign.test(e)||e==="-"?(this._minusSign.lastIndex=0,!0):!1},isDecimalSign:function(e){var t;return(t=this.locale)!==null&&t!==void 0&&t.includes("fr")&&[".",","].includes(e)||this._decimal.test(e)?(this._decimal.lastIndex=0,!0):!1},isDecimalMode:function(){return this.mode==="decimal"},getDecimalCharIndexes:function(e){var t=e.search(this._decimal);this._decimal.lastIndex=0;var i=e.replace(this._prefix,"").trim().replace(/\s/g,"").replace(this._currency,""),u=i.search(this._decimal);return this._decimal.lastIndex=0,{decimalCharIndex:t,decimalCharIndexWithoutPrefix:u}},getCharIndexes:function(e){var t=e.search(this._decimal);this._decimal.lastIndex=0;var i=e.search(this._minusSign);this._minusSign.lastIndex=0;var u=e.search(this._suffix);this._suffix.lastIndex=0;var r=e.search(this._currency);return this._currency.lastIndex=0,{decimalCharIndex:t,minusCharIndex:i,suffixCharIndex:u,currencyCharIndex:r}},insert:function(e,t){var i=arguments.length>2&&arguments[2]!==void 0?arguments[2]:{isDecimalSign:!1,isMinusSign:!1},u=t.search(this._minusSign);if(this._minusSign.lastIndex=0,!(!this.allowMinusSign()&&u!==-1)){var r=this.$refs.input.$el.selectionStart,s=this.$refs.input.$el.selectionEnd,a=this.$refs.input.$el.value.trim(),c=this.getCharIndexes(a),p=c.decimalCharIndex,b=c.minusCharIndex,y=c.suffixCharIndex,o=c.currencyCharIndex,l;if(i.isMinusSign){var v=b===-1;(r===0||r===o+1)&&(l=a,(v||s!==0)&&(l=this.insertText(a,t,0,s)),this.updateValue(e,l,t,"insert"))}else if(i.isDecimalSign)p>0&&r===p?this.updateValue(e,a,t,"insert"):p>r&&p<s?(l=this.insertText(a,t,r,s),this.updateValue(e,l,t,"insert")):p===-1&&this.maxFractionDigits&&(l=this.insertText(a,t,r,s),this.updateValue(e,l,t,"insert"));else{var I=this.numberFormat.resolvedOptions().maximumFractionDigits,g=r!==s?"range-insert":"insert";if(p>0&&r>p){if(r+t.length-(p+1)<=I){var h=o>=r?o-1:y>=r?y:a.length;l=a.slice(0,r)+t+a.slice(r+t.length,h)+a.slice(h),this.updateValue(e,l,t,g)}}else l=this.insertText(a,t,r,s),this.updateValue(e,l,t,g)}}},insertText:function(e,t,i,u){var r=t==="."?t:t.split(".");if(r.length===2){var s=e.slice(i,u).search(this._decimal);return this._decimal.lastIndex=0,s>0?e.slice(0,i)+this.formatValue(t)+e.slice(u):this.formatValue(t)||e}else return u-i===e.length?this.formatValue(t):i===0?t+e.slice(u):u===e.length?e.slice(0,i)+t:e.slice(0,i)+t+e.slice(u)},deleteRange:function(e,t,i){var u;return i-t===e.length?u="":t===0?u=e.slice(i):i===e.length?u=e.slice(0,t):u=e.slice(0,t)+e.slice(i),u},initCursor:function(){var e=this.$refs.input.$el.selectionStart,t=this.$refs.input.$el.value,i=t.length,u=null,r=(this.prefixChar||"").length;t=t.replace(this._prefix,""),e=e-r;var s=t.charAt(e);if(this.isNumeralChar(s))return e+r;for(var a=e-1;a>=0;)if(s=t.charAt(a),this.isNumeralChar(s)){u=a+r;break}else a--;if(u!==null)this.$refs.input.$el.setSelectionRange(u+1,u+1);else{for(a=e;a<i;)if(s=t.charAt(a),this.isNumeralChar(s)){u=a+r;break}else a++;u!==null&&this.$refs.input.$el.setSelectionRange(u,u)}return u||0},onInputClick:function(){var e=this.$refs.input.$el.value;!this.readonly&&e!==R()&&this.initCursor()},isNumeralChar:function(e){return e.length===1&&(this._numeral.test(e)||this._decimal.test(e)||this._group.test(e)||this._minusSign.test(e))?(this.resetRegex(),!0):!1},resetRegex:function(){this._numeral.lastIndex=0,this._decimal.lastIndex=0,this._group.lastIndex=0,this._minusSign.lastIndex=0},updateValue:function(e,t,i,u){var r=this.$refs.input.$el.value,s=null;t!=null&&(s=this.parseValue(t),s=!s&&!this.allowEmpty?0:s,this.updateInput(s,i,u,t),this.handleOnInput(e,r,s))},handleOnInput:function(e,t,i){if(this.isValueChanged(t,i)){var u,r;this.$emit("input",{originalEvent:e,value:i,formattedValue:t}),(u=(r=this.formField).onInput)===null||u===void 0||u.call(r,{originalEvent:e,value:i})}},isValueChanged:function(e,t){if(t===null&&e!==null)return!0;if(t!=null){var i=typeof e=="string"?this.parseValue(e):e;return t!==i}return!1},validateValue:function(e){return e==="-"||e==null?null:this.min!=null&&e<this.min?this.min:this.max!=null&&e>this.max?this.max:e},updateInput:function(e,t,i,u){var r;t=t||"";var s=this.$refs.input.$el.value,a=this.formatValue(e),c=s.length;if(a!==u&&(a=this.concatValues(a,u)),c===0){this.$refs.input.$el.value=a,this.$refs.input.$el.setSelectionRange(0,0);var p=this.initCursor(),b=p+t.length;this.$refs.input.$el.setSelectionRange(b,b)}else{var y=this.$refs.input.$el.selectionStart,o=this.$refs.input.$el.selectionEnd;this.$refs.input.$el.value=a;var l=a.length;if(i==="range-insert"){var v=this.parseValue((s||"").slice(0,y)),I=v!==null?v.toString():"",g=I.split("").join("(".concat(this.groupChar,")?")),h=new RegExp(g,"g");h.test(a);var x=t.split("").join("(".concat(this.groupChar,")?")),$=new RegExp(x,"g");$.test(a.slice(h.lastIndex)),o=h.lastIndex+$.lastIndex,this.$refs.input.$el.setSelectionRange(o,o)}else if(l===c)i==="insert"||i==="delete-back-single"?this.$refs.input.$el.setSelectionRange(o+1,o+1):i==="delete-single"?this.$refs.input.$el.setSelectionRange(o-1,o-1):(i==="delete-range"||i==="spin")&&this.$refs.input.$el.setSelectionRange(o,o);else if(i==="delete-back-single"){var S=s.charAt(o-1),P=s.charAt(o),L=c-l,N=this._group.test(P);N&&L===1?o+=1:!N&&this.isNumeralChar(S)&&(o+=-1*L+1),this._group.lastIndex=0,this.$refs.input.$el.setSelectionRange(o,o)}else if(s==="-"&&i==="insert"){this.$refs.input.$el.setSelectionRange(0,0);var q=this.initCursor(),T=q+t.length+1;this.$refs.input.$el.setSelectionRange(T,T)}else o=o+(l-c),this.$refs.input.$el.setSelectionRange(o,o)}this.$refs.input.$el.setAttribute("aria-valuenow",e),(r=this.$refs.clearIcon)!==null&&r!==void 0&&(r=r.$el)!==null&&r!==void 0&&r.style&&(this.$refs.clearIcon.$el.style.display=U(a)?"none":"block")},concatValues:function(e,t){if(e&&t){var i=t.search(this._decimal);return this._decimal.lastIndex=0,this.suffixChar?i!==-1?e.replace(this.suffixChar,"").split(this._decimal)[0]+t.replace(this.suffixChar,"").slice(i)+this.suffixChar:e:i!==-1?e.split(this._decimal)[0]+t.slice(i):e}return e},getDecimalLength:function(e){if(e){var t=e.split(this._decimal);if(t.length===2)return t[1].replace(this._suffix,"").trim().replace(/\s/g,"").replace(this._currency,"").length}return 0},updateModel:function(e,t){this.writeValue(t,e)},onInputFocus:function(e){this.focused=!0,!this.disabled&&!this.readonly&&this.$refs.input.$el.value!==R()&&this.highlightOnFocus&&e.target.select(),this.$emit("focus",e)},onInputBlur:function(e){var t,i;this.focused=!1;var u=e.target,r=this.validateValue(this.parseValue(u.value));this.$emit("blur",{originalEvent:e,value:u.value}),(t=(i=this.formField).onBlur)===null||t===void 0||t.call(i,e),u.value=this.formatValue(r),u.setAttribute("aria-valuenow",r),this.updateModel(e,r),!this.disabled&&!this.readonly&&this.highlightOnFocus&&Z()},clearTimer:function(){this.timer&&clearTimeout(this.timer)},maxBoundry:function(){return this.d_value>=this.max},minBoundry:function(){return this.d_value<=this.min}},computed:{upButtonListeners:function(){var e=this;return{mousedown:function(i){return e.onUpButtonMouseDown(i)},mouseup:function(i){return e.onUpButtonMouseUp(i)},mouseleave:function(i){return e.onUpButtonMouseLeave(i)},keydown:function(i){return e.onUpButtonKeyDown(i)},keyup:function(i){return e.onUpButtonKeyUp(i)}}},downButtonListeners:function(){var e=this;return{mousedown:function(i){return e.onDownButtonMouseDown(i)},mouseup:function(i){return e.onDownButtonMouseUp(i)},mouseleave:function(i){return e.onDownButtonMouseLeave(i)},keydown:function(i){return e.onDownButtonKeyDown(i)},keyup:function(i){return e.onDownButtonKeyUp(i)}}},formattedValue:function(){var e=!this.d_value&&!this.allowEmpty?0:this.d_value;return this.formatValue(e)},getFormatter:function(){return this.numberFormat},dataP:function(){return Y(F(F({invalid:this.$invalid,fluid:this.$fluid,filled:this.$variant==="filled"},this.size,this.size),this.buttonLayout,this.showButtons&&this.buttonLayout))}},components:{InputText:ne,AngleUpIcon:W,AngleDownIcon:Q,TimesIcon:ee}},we=["data-p"],Ce=["data-p"],xe=["disabled","data-p"],$e=["disabled","data-p"],Se=["disabled","data-p"],Be=["disabled","data-p"];function De(n,e,t,i,u,r){var s=z("InputText"),a=z("TimesIcon");return m(),w("span",d({class:n.cx("root")},n.ptmi("root"),{"data-p":r.dataP}),[K(s,{ref:"input",id:n.inputId,name:n.$formName,role:"spinbutton",class:j([n.cx("pcInputText"),n.inputClass]),style:J(n.inputStyle),defaultValue:r.formattedValue,"aria-valuemin":n.min,"aria-valuemax":n.max,"aria-valuenow":n.d_value,inputmode:n.mode==="decimal"&&!n.minFractionDigits?"numeric":"decimal",disabled:n.disabled,readonly:n.readonly,placeholder:n.placeholder,"aria-labelledby":n.ariaLabelledby,"aria-label":n.ariaLabel,required:n.required,size:n.size,invalid:n.invalid,variant:n.variant,onInput:r.onUserInput,onKeydown:r.onInputKeyDown,onKeypress:r.onInputKeyPress,onPaste:r.onPaste,onClick:r.onInputClick,onFocus:r.onInputFocus,onBlur:r.onInputBlur,pt:n.ptm("pcInputText"),unstyled:n.unstyled,"data-p":r.dataP},null,8,["id","name","class","style","defaultValue","aria-valuemin","aria-valuemax","aria-valuenow","inputmode","disabled","readonly","placeholder","aria-labelledby","aria-label","required","size","invalid","variant","onInput","onKeydown","onKeypress","onPaste","onClick","onFocus","onBlur","pt","unstyled","data-p"]),n.showClear&&n.buttonLayout!=="vertical"?f(n.$slots,"clearicon",{key:0,class:j(n.cx("clearIcon")),clearCallback:r.onClearClick},function(){return[K(a,d({ref:"clearIcon",class:[n.cx("clearIcon")],onClick:r.onClearClick},n.ptm("clearIcon")),null,16,["class","onClick"])]}):B("",!0),n.showButtons&&n.buttonLayout==="stacked"?(m(),w("span",d({key:1,class:n.cx("buttonGroup")},n.ptm("buttonGroup"),{"data-p":r.dataP}),[f(n.$slots,"incrementbutton",{listeners:r.upButtonListeners},function(){return[V("button",d({class:[n.cx("incrementButton"),n.incrementButtonClass]},M(r.upButtonListeners),{disabled:n.disabled,tabindex:-1,"aria-hidden":"true",type:"button"},n.ptm("incrementButton"),{"data-p":r.dataP}),[f(n.$slots,n.$slots.incrementicon?"incrementicon":"incrementbuttonicon",{},function(){return[(m(),D(k(n.incrementIcon||n.incrementButtonIcon?"span":"AngleUpIcon"),d({class:[n.incrementIcon,n.incrementButtonIcon]},n.ptm("incrementIcon"),{"data-pc-section":"incrementicon"}),null,16,["class"]))]})],16,xe)]}),f(n.$slots,"decrementbutton",{listeners:r.downButtonListeners},function(){return[V("button",d({class:[n.cx("decrementButton"),n.decrementButtonClass]},M(r.downButtonListeners),{disabled:n.disabled,tabindex:-1,"aria-hidden":"true",type:"button"},n.ptm("decrementButton"),{"data-p":r.dataP}),[f(n.$slots,n.$slots.decrementicon?"decrementicon":"decrementbuttonicon",{},function(){return[(m(),D(k(n.decrementIcon||n.decrementButtonIcon?"span":"AngleDownIcon"),d({class:[n.decrementIcon,n.decrementButtonIcon]},n.ptm("decrementIcon"),{"data-pc-section":"decrementicon"}),null,16,["class"]))]})],16,$e)]})],16,Ce)):B("",!0),f(n.$slots,"incrementbutton",{listeners:r.upButtonListeners},function(){return[n.showButtons&&n.buttonLayout!=="stacked"?(m(),w("button",d({key:0,class:[n.cx("incrementButton"),n.incrementButtonClass]},M(r.upButtonListeners),{disabled:n.disabled,tabindex:-1,"aria-hidden":"true",type:"button"},n.ptm("incrementButton"),{"data-p":r.dataP}),[f(n.$slots,n.$slots.incrementicon?"incrementicon":"incrementbuttonicon",{},function(){return[(m(),D(k(n.incrementIcon||n.incrementButtonIcon?"span":"AngleUpIcon"),d({class:[n.incrementIcon,n.incrementButtonIcon]},n.ptm("incrementIcon"),{"data-pc-section":"incrementicon"}),null,16,["class"]))]})],16,Se)):B("",!0)]}),f(n.$slots,"decrementbutton",{listeners:r.downButtonListeners},function(){return[n.showButtons&&n.buttonLayout!=="stacked"?(m(),w("button",d({key:0,class:[n.cx("decrementButton"),n.decrementButtonClass]},M(r.downButtonListeners),{disabled:n.disabled,tabindex:-1,"aria-hidden":"true",type:"button"},n.ptm("decrementButton"),{"data-p":r.dataP}),[f(n.$slots,n.$slots.decrementicon?"decrementicon":"decrementbuttonicon",{},function(){return[(m(),D(k(n.decrementIcon||n.decrementButtonIcon?"span":"AngleDownIcon"),d({class:[n.decrementIcon,n.decrementButtonIcon]},n.ptm("decrementIcon"),{"data-pc-section":"decrementicon"}),null,16,["class"]))]})],16,Be)):B("",!0)]})],16,we)}Ie.render=De;export{Ie as s};
