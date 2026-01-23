import{s as v,f as b}from"./index-ChQ6boSU.js";import{B as m,c as p,o as d,p as k,d as y,y as S,x as i,z as h,a as w,t as $,r as l}from"./index-DorqM7ya.js";var P=`
    .p-tag {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        background: dt('tag.primary.background');
        color: dt('tag.primary.color');
        font-size: dt('tag.font.size');
        font-weight: dt('tag.font.weight');
        padding: dt('tag.padding');
        border-radius: dt('tag.border.radius');
        gap: dt('tag.gap');
    }

    .p-tag-icon {
        font-size: dt('tag.icon.size');
        width: dt('tag.icon.size');
        height: dt('tag.icon.size');
    }

    .p-tag-rounded {
        border-radius: dt('tag.rounded.border.radius');
    }

    .p-tag-success {
        background: dt('tag.success.background');
        color: dt('tag.success.color');
    }

    .p-tag-info {
        background: dt('tag.info.background');
        color: dt('tag.info.color');
    }

    .p-tag-warn {
        background: dt('tag.warn.background');
        color: dt('tag.warn.color');
    }

    .p-tag-danger {
        background: dt('tag.danger.background');
        color: dt('tag.danger.color');
    }

    .p-tag-secondary {
        background: dt('tag.secondary.background');
        color: dt('tag.secondary.color');
    }

    .p-tag-contrast {
        background: dt('tag.contrast.background');
        color: dt('tag.contrast.color');
    }
`,B={root:function(e){var t=e.props;return["p-tag p-component",{"p-tag-info":t.severity==="info","p-tag-success":t.severity==="success","p-tag-warn":t.severity==="warn","p-tag-danger":t.severity==="danger","p-tag-secondary":t.severity==="secondary","p-tag-contrast":t.severity==="contrast","p-tag-rounded":t.rounded}]},icon:"p-tag-icon",label:"p-tag-label"},z=m.extend({name:"tag",style:P,classes:B}),E={name:"BaseTag",extends:v,props:{value:null,severity:null,rounded:Boolean,icon:String},style:z,provide:function(){return{$pcTag:this,$parentInstance:this}}};function c(n){"@babel/helpers - typeof";return c=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(e){return typeof e}:function(e){return e&&typeof Symbol=="function"&&e.constructor===Symbol&&e!==Symbol.prototype?"symbol":typeof e},c(n)}function T(n,e,t){return(e=C(e))in n?Object.defineProperty(n,e,{value:t,enumerable:!0,configurable:!0,writable:!0}):n[e]=t,n}function C(n){var e=j(n,"string");return c(e)=="symbol"?e:e+""}function j(n,e){if(c(n)!="object"||!n)return n;var t=n[Symbol.toPrimitive];if(t!==void 0){var a=t.call(n,e);if(c(a)!="object")return a;throw new TypeError("@@toPrimitive must return a primitive value.")}return(e==="string"?String:Number)(n)}var N={name:"Tag",extends:E,inheritAttrs:!1,computed:{dataP:function(){return b(T({rounded:this.rounded},this.severity,this.severity))}}},D=["data-p"];function O(n,e,t,a,o,r){return d(),p("span",i({class:n.cx("root"),"data-p":r.dataP},n.ptmi("root")),[n.$slots.icon?(d(),k(h(n.$slots.icon),i({key:0,class:n.cx("icon")},n.ptm("icon")),null,16,["class"])):n.icon?(d(),p("span",i({key:1,class:[n.cx("icon"),n.icon]},n.ptm("icon")),null,16)):y("",!0),n.value!=null||n.$slots.default?S(n.$slots,"default",{key:2},function(){return[w("span",i({class:n.cx("label")},n.ptm("label")),$(n.value),17)]}):y("",!0)],16,D)}N.render=O;function I(n,e={}){const t=l(null),a=l(null),o=l(!1),r=l(null),u=()=>{try{r.value=new EventSource(n,{withCredentials:!0,...e}),r.value.onopen=()=>{o.value=!0,a.value=null},r.value.onmessage=s=>{try{const g=JSON.parse(s.data);t.value=g}catch{t.value=s.data}},r.value.onerror=s=>{o.value=!1,a.value=s,r.value.readyState!==EventSource.CLOSED&&setTimeout(()=>{u()},5e3)}}catch(s){a.value=s,o.value=!1}},f=()=>{r.value&&(r.value.close(),r.value=null,o.value=!1)};return u(),{data:t,error:a,isConnected:o,disconnect:f,reconnect:u}}export{N as s,I as u};
