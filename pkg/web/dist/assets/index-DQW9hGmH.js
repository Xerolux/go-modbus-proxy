import{a as A}from"./index-DSlDG1eH.js";import{s as S,f as P}from"./index-ChQ6boSU.js";import{B as V,c as l,d as f,o as u,y,p as w,x as p,z as k,t as x,v as D,a as b,F as j,g as $,h as z,q as F,n as g}from"./index-DorqM7ya.js";var L=`
    .p-chip {
        display: inline-flex;
        align-items: center;
        background: dt('chip.background');
        color: dt('chip.color');
        border-radius: dt('chip.border.radius');
        padding-block: dt('chip.padding.y');
        padding-inline: dt('chip.padding.x');
        gap: dt('chip.gap');
    }

    .p-chip-icon {
        color: dt('chip.icon.color');
        font-size: dt('chip.icon.font.size');
        width: dt('chip.icon.size');
        height: dt('chip.icon.size');
    }

    .p-chip-image {
        border-radius: 50%;
        width: dt('chip.image.width');
        height: dt('chip.image.height');
        margin-inline-start: calc(-1 * dt('chip.padding.y'));
    }

    .p-chip:has(.p-chip-remove-icon) {
        padding-inline-end: dt('chip.padding.y');
    }

    .p-chip:has(.p-chip-image) {
        padding-block-start: calc(dt('chip.padding.y') / 2);
        padding-block-end: calc(dt('chip.padding.y') / 2);
    }

    .p-chip-remove-icon {
        cursor: pointer;
        font-size: dt('chip.remove.icon.size');
        width: dt('chip.remove.icon.size');
        height: dt('chip.remove.icon.size');
        color: dt('chip.remove.icon.color');
        border-radius: 50%;
        transition:
            outline-color dt('chip.transition.duration'),
            box-shadow dt('chip.transition.duration');
        outline-color: transparent;
    }

    .p-chip-remove-icon:focus-visible {
        box-shadow: dt('chip.remove.icon.focus.ring.shadow');
        outline: dt('chip.remove.icon.focus.ring.width') dt('chip.remove.icon.focus.ring.style') dt('chip.remove.icon.focus.ring.color');
        outline-offset: dt('chip.remove.icon.focus.ring.offset');
    }
`,T={root:"p-chip p-component",image:"p-chip-image",icon:"p-chip-icon",label:"p-chip-label",removeIcon:"p-chip-remove-icon"},E=V.extend({name:"chip",style:L,classes:T}),R={name:"BaseChip",extends:S,props:{label:{type:[String,Number],default:null},icon:{type:String,default:null},image:{type:String,default:null},removable:{type:Boolean,default:!1},removeIcon:{type:String,default:void 0}},style:E,provide:function(){return{$pcChip:this,$parentInstance:this}}},B={name:"Chip",extends:R,inheritAttrs:!1,emits:["remove"],data:function(){return{visible:!0}},methods:{onKeydown:function(e){(e.key==="Enter"||e.key==="Backspace")&&this.close(e)},close:function(e){this.visible=!1,this.$emit("remove",e)}},computed:{dataP:function(){return P({removable:this.removable})}},components:{TimesCircleIcon:A}},N=["aria-label","data-p"],M=["src"];function W(n,e,i,t,r,o){return r.visible?(u(),l("div",p({key:0,class:n.cx("root"),"aria-label":n.label},n.ptmi("root"),{"data-p":o.dataP}),[y(n.$slots,"default",{},function(){return[n.image?(u(),l("img",p({key:0,src:n.image},n.ptm("image"),{class:n.cx("image")}),null,16,M)):n.$slots.icon?(u(),w(k(n.$slots.icon),p({key:1,class:n.cx("icon")},n.ptm("icon")),null,16,["class"])):n.icon?(u(),l("span",p({key:2,class:[n.cx("icon"),n.icon]},n.ptm("icon")),null,16)):f("",!0),n.label!==null?(u(),l("div",p({key:3,class:n.cx("label")},n.ptm("label")),x(n.label),17)):f("",!0)]}),n.removable?y(n.$slots,"removeicon",{key:0,removeCallback:o.close,keydownCallback:o.onKeydown},function(){return[(u(),w(k(n.removeIcon?"span":"TimesCircleIcon"),p({class:[n.cx("removeIcon"),n.removeIcon],onClick:o.close,onKeydown:o.onKeydown},n.ptm("removeIcon")),null,16,["class","onClick","onKeydown"]))]}):f("",!0)],16,N)):f("",!0)}B.render=W;var U=`
    .p-inputchips {
        display: inline-flex;
    }

    .p-inputchips-input {
        margin: 0;
        list-style-type: none;
        cursor: text;
        overflow: hidden;
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        padding: calc(dt('inputchips.padding.y') / 2) dt('inputchips.padding.x');
        gap: calc(dt('inputchips.padding.y') / 2);
        color: dt('inputchips.color');
        background: dt('inputchips.background');
        border: 1px solid dt('inputchips.border.color');
        border-radius: dt('inputchips.border.radius');
        width: 100%;
        transition:
            background dt('inputchips.transition.duration'),
            color dt('inputchips.transition.duration'),
            border-color dt('inputchips.transition.duration'),
            outline-color dt('inputchips.transition.duration'),
            box-shadow dt('inputchips.transition.duration');
        outline-color: transparent;
        box-shadow: dt('inputchips.shadow');
    }

    .p-inputchips:not(.p-disabled):hover .p-inputchips-input {
        border-color: dt('inputchips.hover.border.color');
    }

    .p-inputchips:not(.p-disabled).p-focus .p-inputchips-input {
        border-color: dt('inputchips.focus.border.color');
        box-shadow: dt('inputchips.focus.ring.shadow');
        outline: dt('inputchips.focus.ring.width') dt('inputchips.focus.ring.style') dt('inputchips.focus.ring.color');
        outline-offset: dt('inputchips.focus.ring.offset');
    }

    .p-inputchips.p-invalid .p-inputchips-input {
        border-color: dt('inputchips.invalid.border.color');
    }

    .p-variant-filled.p-inputchips-input {
        background: dt('inputchips.filled.background');
    }

    .p-inputchips:not(.p-disabled).p-focus .p-variant-filled.p-inputchips-input {
        background: dt('inputchips.filled.focus.background');
    }

    .p-inputchips.p-disabled .p-inputchips-input {
        opacity: 1;
        background: dt('inputchips.disabled.background');
        color: dt('inputchips.disabled.color');
    }

    .p-inputchips-chip.p-chip {
        padding-top: calc(dt('inputchips.padding.y') / 2);
        padding-bottom: calc(dt('inputchips.padding.y') / 2);
        border-radius: dt('inputchips.chip.border.radius');
        transition:
            background dt('inputchips.transition.duration'),
            color dt('inputchips.transition.duration');
    }

    .p-inputchips-chip-item.p-focus .p-inputchips-chip {
        background: dt('inputchips.chip.focus.background');
        color: dt('inputchips.chip.focus.color');
    }

    .p-inputchips-input:has(.p-inputchips-chip) {
        padding-left: calc(dt('inputchips.padding.y') / 2);
        padding-right: calc(dt('inputchips.padding.y') / 2);
    }

    .p-inputchips-input-item {
        flex: 1 1 auto;
        display: inline-flex;
        padding-top: calc(dt('inputchips.padding.y') / 2);
        padding-bottom: calc(dt('inputchips.padding.y') / 2);
    }

    .p-inputchips-input-item input {
        border: 0 none;
        outline: 0 none;
        background: transparent;
        margin: 0;
        padding: 0;
        box-shadow: none;
        border-radius: 0;
        width: 100%;
        font-family: inherit;
        font-feature-settings: inherit;
        font-size: 1rem;
        color: inherit;
    }

    .p-inputchips-input-item input::placeholder {
        color: dt('inputchips.placeholder.color');
    }
`,q={root:function(e){var i=e.instance,t=e.props;return["p-inputchips p-component p-inputwrapper",{"p-disabled":t.disabled,"p-invalid":t.invalid,"p-focus":i.focused,"p-inputwrapper-filled":t.modelValue&&t.modelValue.length||i.inputValue&&i.inputValue.length,"p-inputwrapper-focus":i.focused}]},input:function(e){var i=e.props,t=e.instance;return["p-inputchips-input",{"p-variant-filled":i.variant?i.variant==="filled":t.$primevue.config.inputStyle==="filled"||t.$primevue.config.inputVariant==="filled"}]},chipItem:function(e){var i=e.state,t=e.index;return["p-inputchips-chip-item",{"p-focus":i.focusedIndex===t}]},pcChip:"p-inputchips-chip",chipIcon:"p-inputchips-chip-icon",inputItem:"p-inputchips-input-item"},H=V.extend({name:"inputchips",style:U,classes:q}),G={name:"BaseInputChips",extends:S,props:{modelValue:{type:Array,default:null},max:{type:Number,default:null},separator:{type:[String,Object],default:null},addOnBlur:{type:Boolean,default:null},allowDuplicate:{type:Boolean,default:!0},placeholder:{type:String,default:null},variant:{type:String,default:null},invalid:{type:Boolean,default:!1},disabled:{type:Boolean,default:!1},inputId:{type:String,default:null},inputClass:{type:[String,Object],default:null},inputStyle:{type:Object,default:null},inputProps:{type:null,default:null},removeTokenIcon:{type:String,default:void 0},chipIcon:{type:String,default:void 0},ariaLabelledby:{type:String,default:null},ariaLabel:{type:String,default:null}},style:H,provide:function(){return{$pcInputChips:this,$parentInstance:this}}};function m(n){return Y(n)||X(n)||Q(n)||J()}function J(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function Q(n,e){if(n){if(typeof n=="string")return v(n,e);var i={}.toString.call(n).slice(8,-1);return i==="Object"&&n.constructor&&(i=n.constructor.name),i==="Map"||i==="Set"?Array.from(n):i==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(i)?v(n,e):void 0}}function X(n){if(typeof Symbol<"u"&&n[Symbol.iterator]!=null||n["@@iterator"]!=null)return Array.from(n)}function Y(n){if(Array.isArray(n))return v(n)}function v(n,e){(e==null||e>n.length)&&(e=n.length);for(var i=0,t=Array(e);i<e;i++)t[i]=n[i];return t}var K={name:"InputChips",extends:G,inheritAttrs:!1,emits:["update:modelValue","add","remove","focus","blur"],data:function(){return{inputValue:null,focused:!1,focusedIndex:null}},mounted:function(){console.warn("Deprecated since v4. Use AutoComplete component instead with its typeahead property.")},methods:{onWrapperClick:function(){this.$refs.input.focus()},onInput:function(e){this.inputValue=e.target.value,this.focusedIndex=null},onFocus:function(e){this.focused=!0,this.focusedIndex=null,this.$emit("focus",e)},onBlur:function(e){this.focused=!1,this.focusedIndex=null,this.addOnBlur&&this.addItem(e,e.target.value,!1),this.$emit("blur",e)},onKeyDown:function(e){var i=e.target.value;switch(e.code){case"Backspace":i.length===0&&this.modelValue&&this.modelValue.length>0&&(this.focusedIndex!==null?this.removeItem(e,this.focusedIndex):this.removeItem(e,this.modelValue.length-1));break;case"Enter":case"NumpadEnter":i&&i.trim().length&&!this.maxedOut&&this.addItem(e,i,!0);break;case"ArrowLeft":i.length===0&&this.modelValue&&this.modelValue.length>0&&this.$refs.container.focus();break;case"ArrowRight":e.stopPropagation();break;default:this.separator&&(this.separator===e.key||e.key.match(this.separator))&&this.addItem(e,i,!0);break}},onPaste:function(e){var i=this;if(this.separator){var t=this.separator.replace("\\n",`
`).replace("\\r","\r").replace("\\t","	"),r=(e.clipboardData||window.clipboardData).getData("Text");if(r){var o=this.modelValue||[],d=r.split(t);d=d.filter(function(a){return i.allowDuplicate||o.indexOf(a)===-1}),o=[].concat(m(o),m(d)),this.updateModel(e,o,!0)}}},onContainerFocus:function(){this.focused=!0},onContainerBlur:function(){this.focusedIndex=-1,this.focused=!1},onContainerKeyDown:function(e){switch(e.code){case"ArrowLeft":this.onArrowLeftKeyOn(e);break;case"ArrowRight":this.onArrowRightKeyOn(e);break;case"Backspace":this.onBackspaceKeyOn(e);break}},onArrowLeftKeyOn:function(){this.inputValue.length===0&&this.modelValue&&this.modelValue.length>0&&(this.focusedIndex=this.focusedIndex===null?this.modelValue.length-1:this.focusedIndex-1,this.focusedIndex<0&&(this.focusedIndex=0))},onArrowRightKeyOn:function(){this.inputValue.length===0&&this.modelValue&&this.modelValue.length>0&&(this.focusedIndex===this.modelValue.length-1?(this.focusedIndex=null,this.$refs.input.focus()):this.focusedIndex++)},onBackspaceKeyOn:function(e){this.focusedIndex!==null&&this.removeItem(e,this.focusedIndex)},updateModel:function(e,i,t){var r=this;this.$emit("update:modelValue",i),this.$emit("add",{originalEvent:e,value:i}),this.$refs.input.value="",this.inputValue="",setTimeout(function(){r.maxedOut&&(r.focused=!1)},0),t&&e.preventDefault()},addItem:function(e,i,t){if(i&&i.trim().length){var r=this.modelValue?m(this.modelValue):[];(this.allowDuplicate||r.indexOf(i)===-1)&&(r.push(i),this.updateModel(e,r,t))}},removeItem:function(e,i){if(!this.disabled){var t=m(this.modelValue),r=t.splice(i,1);this.focusedIndex=null,this.$refs.input.focus(),this.$emit("update:modelValue",t),this.$emit("remove",{originalEvent:e,value:r})}}},computed:{maxedOut:function(){return this.max&&this.modelValue&&this.max===this.modelValue.length},focusedOptionId:function(){return this.focusedIndex!==null?"".concat(this.$id,"_inputchips_item_").concat(this.focusedIndex):null}},components:{Chip:B}};function h(n){"@babel/helpers - typeof";return h=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(e){return typeof e}:function(e){return e&&typeof Symbol=="function"&&e.constructor===Symbol&&e!==Symbol.prototype?"symbol":typeof e},h(n)}function C(n,e){var i=Object.keys(n);if(Object.getOwnPropertySymbols){var t=Object.getOwnPropertySymbols(n);e&&(t=t.filter(function(r){return Object.getOwnPropertyDescriptor(n,r).enumerable})),i.push.apply(i,t)}return i}function O(n){for(var e=1;e<arguments.length;e++){var i=arguments[e]!=null?arguments[e]:{};e%2?C(Object(i),!0).forEach(function(t){Z(n,t,i[t])}):Object.getOwnPropertyDescriptors?Object.defineProperties(n,Object.getOwnPropertyDescriptors(i)):C(Object(i)).forEach(function(t){Object.defineProperty(n,t,Object.getOwnPropertyDescriptor(i,t))})}return n}function Z(n,e,i){return(e=_(e))in n?Object.defineProperty(n,e,{value:i,enumerable:!0,configurable:!0,writable:!0}):n[e]=i,n}function _(n){var e=nn(n,"string");return h(e)=="symbol"?e:e+""}function nn(n,e){if(h(n)!="object"||!n)return n;var i=n[Symbol.toPrimitive];if(i!==void 0){var t=i.call(n,e);if(h(t)!="object")return t;throw new TypeError("@@toPrimitive must return a primitive value.")}return(e==="string"?String:Number)(n)}var en=["aria-labelledby","aria-label","aria-activedescendant"],tn=["id","aria-label","aria-setsize","aria-posinset","data-p-focused"],on=["id","disabled","placeholder","aria-invalid"];function rn(n,e,i,t,r,o){var d=D("Chip");return u(),l("div",p({class:n.cx("root")},n.ptmi("root")),[b("ul",p({ref:"container",class:n.cx("input"),tabindex:"-1",role:"listbox","aria-orientation":"horizontal","aria-labelledby":n.ariaLabelledby,"aria-label":n.ariaLabel,"aria-activedescendant":r.focused?o.focusedOptionId:void 0,onClick:e[5]||(e[5]=function(a){return o.onWrapperClick()}),onFocus:e[6]||(e[6]=function(){return o.onContainerFocus&&o.onContainerFocus.apply(o,arguments)}),onBlur:e[7]||(e[7]=function(){return o.onContainerBlur&&o.onContainerBlur.apply(o,arguments)}),onKeydown:e[8]||(e[8]=function(){return o.onContainerKeyDown&&o.onContainerKeyDown.apply(o,arguments)})},n.ptm("input")),[(u(!0),l(j,null,$(n.modelValue,function(a,s){return u(),l("li",p({key:"".concat(s,"_").concat(a),id:n.$id+"_inputchips_item_"+s,role:"option",class:n.cx("chipItem",{index:s}),"aria-label":a,"aria-selected":!0,"aria-setsize":n.modelValue.length,"aria-posinset":s+1},{ref_for:!0},n.ptm("chipItem"),{"data-p-focused":r.focusedIndex===s}),[y(n.$slots,"chip",{class:g(n.cx("pcChip")),index:s,value:a,removeCallback:function(c){return n.removeOption(c,s)}},function(){return[z(d,{class:g(n.cx("pcChip")),label:a,removeIcon:n.chipIcon||n.removeTokenIcon,removable:"",unstyled:n.unstyled,onRemove:function(c){return o.removeItem(c,s)},pt:n.ptm("pcChip")},{removeicon:F(function(){return[y(n.$slots,n.$slots.chipicon?"chipicon":"removetokenicon",{class:g(n.cx("chipIcon")),index:s,removeCallback:function(c){return o.removeItem(c,s)}})]}),_:2},1032,["class","label","removeIcon","unstyled","onRemove","pt"])]})],16,tn)}),128)),b("li",p({class:n.cx("inputItem"),role:"option"},n.ptm("inputItem")),[b("input",p({ref:"input",id:n.inputId,type:"text",class:n.inputClass,style:n.inputStyle,disabled:n.disabled||o.maxedOut,placeholder:n.placeholder,"aria-invalid":n.invalid||void 0,onFocus:e[0]||(e[0]=function(a){return o.onFocus(a)}),onBlur:e[1]||(e[1]=function(a){return o.onBlur(a)}),onInput:e[2]||(e[2]=function(){return o.onInput&&o.onInput.apply(o,arguments)}),onKeydown:e[3]||(e[3]=function(a){return o.onKeyDown(a)}),onPaste:e[4]||(e[4]=function(a){return o.onPaste(a)})},O(O({},n.inputProps),n.ptm("inputItemField"))),null,16,on)],16)],16,en)],16)}K.render=rn;var un={name:"Chips",extends:K,mounted:function(){console.warn("Deprecated since v4. Use InputChips component instead.")}};export{un as s};
